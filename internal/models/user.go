package models

import "github.com/jackc/pgx/v5"

type TwoFA string

const (
	TwoFA_Disable TwoFA = "disable"
	TwoFA_TOTP    TwoFA = "totp"
)

type User struct {
	ID int

	Email        string
	PasswordHash string
	TwoFaType    TwoFA
}

func (m *User) FromRow(row pgx.Row) error {
	return row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.TwoFaType)
}

func UserFactory() *User {
	return &User{}
}
