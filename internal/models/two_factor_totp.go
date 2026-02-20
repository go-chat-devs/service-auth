package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TwoFactorTOTP struct {
	userUID     uuid.UUID
	totp_secret string
}

func (t *TwoFactorTOTP) FromRow(row pgx.Row) error {
	return row.Scan(&t.userUID, &t.totp_secret)
}

func TwoFactorTOTPFactory() *TwoFactorTOTP {
	return &TwoFactorTOTP{}
}
