package twofactortotp

import (
	"context"
	"log/slog"

	"github.com/go-chat-devs/service-auth/internal/models"
	"github.com/go-chat-devs/service-auth/internal/scanner"
	"github.com/go-chat-devs/service-auth/internal/storage/db"
	"github.com/go-chat-devs/service-auth/internal/tagger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var tag = tagger.Tagger("storage-2fa-totp")

type Storage struct {
	db db.DBTX
}

func New(db db.DBTX) *Storage {
	return &Storage{db: db}
}

func (s *Storage) WithTX(tx pgx.Tx) *Storage {
	return &Storage{db: tx}
}

func (s *Storage) Insert(ctx context.Context, userUID uuid.UUID, secret string) error {
	const sql = `INSERT INTO two_factor_totp(user_uid, secret) VALUES($1, $2)`
	_, err := s.db.Exec(ctx, sql, userUID, secret)
	if err != nil {
		slog.Error(tag("insert error: %v", err))
	}
	return err
}

func (s *Storage) Select(ctx context.Context, userUID uuid.UUID) (*models.TwoFactorTOTP, error) {
	const sql = `SELECT * FROM two_factor_totp WHERE user_uid = $1`
	row := s.db.QueryRow(ctx, sql, userUID)
	res, err := scanner.Row(row, models.TwoFactorTOTPFactory)
	if err != nil {
		slog.Error(tag("select error: %v", err))
		return nil, err
	}
	return res, nil
}

func (s *Storage) Delete(ctx context.Context, userUID uuid.UUID) error {
	const sql = `DELETE FROM two_factor_totp WHERE user_uid=$1`
	_, err := s.db.Exec(ctx, sql, userUID)
	if err != nil {
		slog.Error(tag("delete error: %v", err))
	}
	return err
}
