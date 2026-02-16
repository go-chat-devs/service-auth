package models

import "github.com/jackc/pgx/v5"

type two_fa_type string

const (
	disable two_fa_type = "disable"
	totp    two_fa_type = "totp"
)

type User struct {
	ID int

	Email        string
	PasswordHash string
	TwoFaType    two_fa_type
}

func (m *User) FromRow(row pgx.Row) error {
	return row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.TwoFaType)
}

func UserFactory() *User {
	return &User{}
}
