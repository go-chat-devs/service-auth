package models

import (
	"github.com/jackc/pgx/v5"
)

type TwoFactorTOTP struct {
	UserID int
	Secret string
}

func (t *TwoFactorTOTP) FromRow(row pgx.Row) error {
	return row.Scan(&t.UserID, &t.Secret)
}

func TwoFactorTOTPFactory() *TwoFactorTOTP {
	return &TwoFactorTOTP{}
}
