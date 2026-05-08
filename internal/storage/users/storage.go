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

func (s *Storage) Insert(ctx context.Context, email string, passwordHash []byte, twoFaType models.TwoFA) error {
	sql := "INSERT INTO auth.users(email, password_hash, two_fa_type) VALUES($1,$2,$3)"
	_, err := s.db.Exec(ctx, sql, email, passwordHash, twoFaType)
	if err != nil {
		slog.Error(tag("insert error: %v", err))
	}
	return err
}

func (s *Storage) Delete(ctx context.Context, userID int) error {
	sql := "DELETE FROM auth.users WHERE id=$1"
	_, err := s.db.Exec(ctx, sql, userID)
	if err != nil {
		slog.Error(tag("delete error: %v", err))
	}
	return err
}

func (s *Storage) UpdateEmail(ctx context.Context, userID int, newEmail string) error {
	sql := "UPDATE auth.users SET email=$1 WHERE id=$2"
	_, err := s.db.Exec(ctx, sql, newEmail, userID)
	if err != nil {
		slog.Error(tag("update email error: %v", err))
	}
	return err
}

func (s *Storage) UpdatePassword(ctx context.Context, userID int, newHash []byte) error {
	sql := "UPDATE auth.users SET password_hash=$1 WHERE id=$2"
	_, err := s.db.Exec(ctx, sql, newHash, userID)
	if err != nil {
		slog.Error(tag("update password error: %v", err))
	}
	return err
}

func (s *Storage) Update2FA(ctx context.Context, userID int, newType models.TwoFA) error {
	sql := "UPDATE auth.users SET two_fa_type=$1 WHERE id=$2"
	_, err := s.db.Exec(ctx, sql, newType, userID)
	if err != nil {
		slog.Error(tag("update 2fa error: %v", err))
	}
	return err
}

func (s *Storage) Select(ctx context.Context, userID int) (*models.User, error) {
	sql := "SELECT * FROM auth.users WHERE id=$1"
	row := s.db.QueryRow(ctx, sql, userID)
	res, err := scanner.Row(row, models.UserFactory)
	if err != nil {
		slog.Error(tag("select error: %v", err))
		return nil, err
	}
	return res, nil
}

func (s *Storage) SelectEmail(ctx context.Context, email string) (*models.User, error) {
	sql := "SELECT * FROM auth.users WHERE email=$1"
	row := s.db.QueryRow(ctx, sql, email)
	res, err := scanner.Row(row, models.UserFactory)
	if err != nil {
		slog.Error(tag("select email error: %v", err))
		return nil, err
	}
	return res, nil
}
