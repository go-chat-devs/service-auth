package custom_errors

import "errors"

var(
	ErrUserExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTOTPAlreadySetup = errors.New("2fa already setup")
	ErrTOTPExpToken = errors.New("expired token")
	ErrInvalidCode = errors.New("invalid code")
)

type Require2FA_TOTP struct {
}

func (e *Require2FA_TOTP) Error() string {
	return "2fa required"
}
