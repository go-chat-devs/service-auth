package models

import (
	"github.com/go-chat-devs/service-auth/internal/token"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	SessionKey token.Token
	UserID     int
}

func (m *Session) FromRow(row pgx.Row) error {
	return row.Scan(&m.SessionKey, &m.UserID)
}

func SessionFactory() *Session {
	return &Session{}
}
