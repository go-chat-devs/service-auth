package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	SessionKey string
	UID        uuid.UUID
}

func (m *Session) FromRow(row pgx.Row) error {
	return row.Scan(&m.SessionKey, &m.UID)
}

func SessionFactory() *Session {
	return &Session{}
}
