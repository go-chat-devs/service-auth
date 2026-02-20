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
	"github.com/pquerna/otp/totp"
)

type Storage struct {
	gw   *gateway
	pool *pgxpool.Pool

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
		gw:            newGateway(ctx),
		pool:          pool,
		users:         users.New(pool),
		sessions:      sessions.New(pool),
		twoFactorTotp: twofactortotp.New(pool),
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
			return errors.New("invalid password")
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
			tok = s.gw.Store(user.ID)
			err = &custom_errors.Require2FA_TOTP{}
		}
		return nil
	})
	return
}
func (s *Storage) ValidateTOTP(ctx context.Context, token2fa token.Token, code string) (sessionKey token.Token, err error) {
	userId, ok := s.gw.Get(token2fa)
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
		Totp := s.twoFactorTotp.WithTX(tx)
		entry, err := Totp.Select(ctx, usr.ID)
		if !totp.Validate(code, entry.Secret) {
			return errors.New("invalid code")
		}
		Sessions := s.sessions.WithTX(tx)
		sess, err := Sessions.Insert(ctx, userId)
		sessionKey = sess.SessionKey
		return err
	})
	return
}
func (s *Storage) GetUserID(ctx context.Context, sessionKey token.Token) (userId int, err error) {
	sess, err := s.sessions.Select(ctx, sessionKey)
	userId = sess.UserID
	return
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
			return errors.New("invalid password")
		}
		return Users.Delete(ctx, user.ID)
	})
}
