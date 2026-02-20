package users

import (
	"context"
	"log/slog"

	"github.com/go-chat-devs/service-auth/internal/models"
	"github.com/go-chat-devs/service-auth/internal/scanner"
	"github.com/go-chat-devs/service-auth/internal/storage/db"
	"github.com/go-chat-devs/service-auth/internal/tagger"
	"github.com/jackc/pgx/v5"
)

var tag = tagger.Tagger("storage-users")

type Storage struct {
	db db.DBTX
}

func New(db db.DBTX) *Storage {
	return &Storage{db: db}
}

func (s *Storage) WithTX(tx pgx.Tx) *Storage {
	return &Storage{db: tx}
}

func (s *Storage) Insert(ctx context.Context, email string, passwordHash string, twoFaType models.TwoFA) error {
	sql := "INSERT INTO auth.users (email,password_hash,two_fa_type) VALUES ($1,$2,$3)"
	_, err := s.db.Exec(ctx, sql, email, passwordHash, twoFaType)
	if err != nil {
		slog.Error(tag("insert error: %v", err))
	}
	return err
}

func (s *Storage) Delete(ctx context.Context, id int) error {
	sql := "INSERT * FROM auth.users WHERE id=$1;"
	_, err := s.db.Exec(ctx, sql, id)
	if err != nil {
		slog.Error(tag("delete error: %v", err))
	}
	return err
}

func (s *Storage) UpdateEmail(ctx context.Context, id int, newEmail string) error {
	sql := "UPDATE auth.users SET email=$1 WHERE id=$2;"
	_, err := s.db.Exec(ctx, sql, newEmail, id)
	if err != nil {
		slog.Error(tag("update email error: %v", err))
	}
	return err
}

func (s *Storage) UpdatePassword(ctx context.Context, id int, newPass string) error {
	sql := "UPDATE auth.users SET password_hash=$1 WHERE id=$2;"
	_, err := s.db.Exec(ctx, sql, newPass, id)
	if err != nil {
		slog.Error(tag("update password error: %v", err))
	}
	return err
}

func (s *Storage) Update2FA(ctx context.Context, id int, newType models.TwoFA) error {
	sql := "UPDATE auth.users SET two_fa_type=$1 WHERE id=$2;"
	_, err := s.db.Exec(ctx, sql, newType, id)
	if err != nil {
		slog.Error(tag("update 2fa error: %v", err))
	}
	return err
}

func (s *Storage) Select(ctx context.Context, id int) (*models.User, error) {
	sql := "SELECT * FROM auth.users WHERE id=$1"
	row := s.db.QueryRow(ctx, sql, id)
	res, err := scanner.Row(row, models.UserFactory)
	if err != nil {
		slog.Error(tag("select error: %v", err))
		return nil, err
	}
	return res, nil
}
