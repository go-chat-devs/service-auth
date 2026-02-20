package custom_errors

type Require2FA_TOTP struct {
}

func (e *Require2FA_TOTP) Error() string {
	return "2fa required"
}
