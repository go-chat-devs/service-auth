package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	custom_errors "github.com/go-chat-devs/service-auth/internal/errors"
	"github.com/go-chat-devs/service-auth/internal/token"
	"github.com/pquerna/otp"
)

type Auth struct {
	log            *slog.Logger
	usrManager     UserManager
	totpManager    TOTPManager
	sessionManager SessionManager
	tokenTTL       time.Duration
}

type UserManager interface {
	CreateUser(
		ctx context.Context,
		email string,
		password string,
	) (err error)
	AuthenticateUser(
		ctx context.Context,
		email string,
		password string,
	) (tok token.Token, err error)

	GetUserID(
		ctx context.Context,
		sessionKey token.Token,
	) (userId int, err error)
	DeleteUser(
		ctx context.Context,
		sessionKey token.Token,
		password string,
	) (err error)
}

type SessionManager interface {
	DeleteSession(ctx context.Context, sessionKey token.Token) error
	DeleteAllSessions(ctx context.Context, userID int) error
}

type TOTPManager interface {
	SetupTOTP_step1(ctx context.Context,
		sessionKey token.Token,
	) (res struct {
		Key   *otp.Key
		Token token.Token
	}, err error)
	SetupTOTP_step2(ctx context.Context,
		sessionKey token.Token,
		setup2fa token.Token,
		code string,
	) (err error)
	ValidateTOTP(
		ctx context.Context,
		token2fa token.Token,
		code string,
	) (sessionKey token.Token, err error)
}

func New(
	log *slog.Logger,
	usrManager UserManager,
	totpManager TOTPManager,
	sessionManager SessionManager,
	tokenTTl time.Duration,
) *Auth {
	return &Auth{
		log:            log,
		usrManager:     usrManager,
		totpManager:    totpManager,
		sessionManager: sessionManager,
		tokenTTL:       tokenTTl,
	}
}

func (a *Auth) Register(ctx context.Context, email, password string) (err error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op))
	err = a.usrManager.CreateUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, custom_errors.ErrUserExists) {
			log.Warn("user already exists")
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to save user", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registred")
	return nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (token token.Token, err error) {
	const op = "auth.LoginUser"
	log := a.log.With(slog.String("op", op))

	token, err = a.usrManager.AuthenticateUser(ctx, email, password)
	if err != nil {

		if errors.Is(err, custom_errors.ErrInvalidCredentials) {
			log.Error("invalid credentials", slog.String("err", err.Error()))
			return [32]byte{}, fmt.Errorf("%s: %w", op, err)
		}

		var require2FA *custom_errors.Require2FA_TOTP
		if errors.As(err, &require2FA) {
			log.Info("2FA required", slog.String("email", email))
			return [32]byte{}, err
		}

		log.Error("authentication failed", slog.String("err", err.Error()))
		return [32]byte{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged successfully")
	return token, nil
}

func (a *Auth) SetupTOTP(
	ctx context.Context,
	sessionkey token.Token,
) (res struct {
	Key   *otp.Key
	Token token.Token
}, err error) {
	const op = "auth.SetupTotp"
	log := a.log.With(slog.String("op", op))
	res, err = a.totpManager.SetupTOTP_step1(ctx, sessionkey)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTOTPAlreadySetup) {
			log.Warn("2fa already setup")
			return struct {
				Key   *otp.Key
				Token token.Token
			}{}, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("setup totp failed", slog.String("err", err.Error()))
		return struct {
			Key   *otp.Key
			Token token.Token
		}{}, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("setup 2fa succesfully")
	return
}

func (a *Auth) SetupTOTPValidate(
	ctx context.Context,
	sessionkey token.Token,
	setup2fa token.Token,
	code string,
) (err error) {
	const op = "auth.SetupTotpValidate"
	log := a.log.With(slog.String("op", op))
	err = a.totpManager.SetupTOTP_step2(ctx, sessionkey, setup2fa, code)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTOTPExpToken) {
			log.Error("expired token")
			return fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, custom_errors.ErrInvalidCode) {
			log.Error("invalid code")
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("setup 2fa validate failed")
		return fmt.Errorf("%s: %w", op, err)
	}
	return
}

func (a *Auth) ValidateTOTP(
	ctx context.Context,
	token2fa token.Token,
	code string,
) (sessionkey token.Token, err error) {
	const op = "auth.ValidateTOTP"
	log := a.log.With(slog.String("op", op))
	sessionkey, err = a.totpManager.ValidateTOTP(ctx, token2fa, code)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTOTPExpToken) {
			log.Error("expired token")
			return [32]byte{}, fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, custom_errors.ErrInvalidCode) {
			log.Error("invalid code")
			return [32]byte{}, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("2fa validation failed")
		return [32]byte{}, fmt.Errorf("%s: %w", op, err)
	}
	return
}

func (a *Auth) Logout(
	ctx context.Context,
	sessionkey token.Token,
) (err error) {
	const op = "auth.Logout"
	log := a.log.With(slog.String("op", op))
	err = a.sessionManager.DeleteSession(ctx, sessionkey)
	if err != nil {
		log.Error("logout failed")
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.Info("logout succesfully")
	return
}

func (a *Auth) DeleteAccount(ctx context.Context, sessionkey token.Token, password string) (err error) {
	const op = "auth.DeleteAccount"
	log := a.log.With(slog.String("op", op))
	err = a.DeleteAccount(ctx, sessionkey, password)
	if err != nil {
		if errors.Is(err, custom_errors.ErrInvalidCredentials) {
			log.Error("invalid credentials", slog.String("err", err.Error()))
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to delete user account", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	return
}
