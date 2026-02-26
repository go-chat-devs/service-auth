package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	custom_errors "github.com/go-chat-devs/service-auth/internal/errors"
	"github.com/go-chat-devs/service-auth/internal/models"
	"github.com/go-chat-devs/service-auth/internal/storage/db"
	"github.com/go-chat-devs/service-auth/internal/storage/pwdgen"
	"github.com/go-chat-devs/service-auth/internal/storage/sessions"
	twofactortotp "github.com/go-chat-devs/service-auth/internal/storage/two_factor_totp"
	"github.com/go-chat-devs/service-auth/internal/storage/users"
	"github.com/go-chat-devs/service-auth/internal/token"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type Storage struct {
	gwTotpValidate *gateway[int]
	gwTotpSetup    *gateway[*otp.Key]
	pool           *pgxpool.Pool

	users         *users.Storage
	sessions      *sessions.Storage
	twoFactorTotp *twofactortotp.Storage
}

func New(ctx context.Context) (*Storage, error) {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		slog.Error(fmt.Sprintf("error parsing connection config: %v", err))
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("error creating connection pool: %v", err))
		return nil, err
	}

	return &Storage{
		gwTotpValidate: newGateway[int](ctx),
		gwTotpSetup:    newGateway[*otp.Key](ctx),
		pool:           pool,
		users:          users.New(pool),
		sessions:       sessions.New(pool),
		twoFactorTotp:  twofactortotp.New(pool),
	}, nil
}

func (s *Storage) CreateUser(ctx context.Context, email, password string) error {
	hash := pwdgen.Generate([]byte(password))
	return s.users.Insert(ctx, email, hash, models.TwoFA_Disable)
}
func (s *Storage) AuthenticateUser(ctx context.Context, email, password string) (tok token.Token, err error) {
	err = db.Transaction(ctx, s.pool, func(tx pgx.Tx) error {
		Users := s.users.WithTX(tx)
		user, err := Users.SelectEmail(ctx, email)
		if err != nil {
			return err
		}
		if !pwdgen.Check([]byte(password), user.PasswordHash) {
			return errors.New("invalid credentials")
		}
		switch user.TwoFaType {
		case models.TwoFA_Disable:
			Sessions := s.sessions.WithTX(tx)
			sess, err := Sessions.Insert(ctx, user.ID)
			if err != nil {
				return err
			}
			tok = sess.SessionKey
			err = nil
		case models.TwoFA_TOTP:
			tok = s.gwTotpValidate.Store(user.ID)
			err = &custom_errors.Require2FA_TOTP{}
		}
		return nil
	})
	return
}
func (s *Storage) SetupTOTP_step1(ctx context.Context, sessionKey token.Token) (res struct {
	Key   *otp.Key
	Token token.Token
}, err error) {
	err = db.Transaction(ctx, s.pool, func(tx pgx.Tx) error {
		Sessions := s.sessions.WithTX(tx)
		sess, err := Sessions.Select(ctx, sessionKey)
		if err != nil {
			return err
		}
		Users := s.users.WithTX(tx)
		user, err := Users.Select(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if user.TwoFaType != models.TwoFA_Disable {
			return errors.New("2fa already setup")
		}
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Go chat",
			AccountName: user.Email,
		})
		if err != nil {
			return err
		}
		tok := s.gwTotpSetup.Store(key)
		res = struct {
			Key   *otp.Key
			Token token.Token
		}{
			Key:   key,
			Token: tok,
		}
		return nil
	})
	return
}
func (s *Storage) SetupTOTP_step2(ctx context.Context, sessionKey, setup2fa token.Token, code string) (err error) {
	key, ok := s.gwTotpSetup.Get(setup2fa)
	if !ok {
		return errors.New("expired token")
	}
	if !totp.Validate(code, key.Secret()) {
		return errors.New("invalid code")
	}
	err = db.Transaction(ctx, s.pool, func(tx pgx.Tx) error {
		Sessions := s.sessions.WithTX(tx)
		sess, err := Sessions.Select(ctx, sessionKey)
		if err != nil {
			return err
		}
		Users := s.users.WithTX(tx)
		if err = Users.Update2FA(ctx, sess.UserID, models.TwoFA_TOTP); err != nil {
			return err
		}
		TOTP := s.twoFactorTotp.WithTX(tx)
		return TOTP.Insert(ctx, sess.UserID, key.Secret())
	})
	if err == nil {
		s.gwTotpSetup.Erase(setup2fa)
	}
	return err
}
func (s *Storage) ValidateTOTP(ctx context.Context, token2fa token.Token, code string) (sessionKey token.Token, err error) {
	userId, ok := s.gwTotpValidate.Get(token2fa)
	if !ok {
		err = errors.New("expired token")
		return
	}
	err = db.Transaction(ctx, s.pool, func(tx pgx.Tx) error {
		Users := s.users.WithTX(tx)
		usr, err := Users.Select(ctx, userId)
		if err != nil {
			return err
		}
		TOTP := s.twoFactorTotp.WithTX(tx)
		entry, err := TOTP.Select(ctx, usr.ID)
		if err != nil {
			return err
		}
		if !totp.Validate(code, entry.Secret) {
			return errors.New("invalid code")
		}
		Sessions := s.sessions.WithTX(tx)
		sess, err := Sessions.Insert(ctx, userId)
		if err != nil {
			return err
		}
		sessionKey = sess.SessionKey
		return nil
	})
	if err == nil {
		s.gwTotpValidate.Erase(token2fa)
	}
	return
}
func (s *Storage) GetUserID(ctx context.Context, sessionKey token.Token) (userId int, err error) {
	sess, err := s.sessions.Select(ctx, sessionKey)
	return sess.UserID, err
}
func (s *Storage) DeleteUser(ctx context.Context, sessionKey token.Token, password string) error {
	return db.Transaction(ctx, s.pool, func(tx pgx.Tx) error {
		Sessions := s.sessions.WithTX(tx)
		sess, err := Sessions.Select(ctx, sessionKey)
		if err != nil {
			return err
		}
		Users := s.users.WithTX(tx)
		user, err := Users.Select(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if !pwdgen.Check([]byte(password), user.PasswordHash) {
			return errors.New("invalid credentials")
		}
		return Users.Delete(ctx, user.ID)
	})
}
