package sessions

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

var tag = tagger.Tagger("storage-sessions")

type Storage struct {
	db db.DBTX
}

func New(db db.DBTX) *Storage {
	return &Storage{db: db}
}

func (s *Storage) WithTX(tx pgx.Tx) *Storage {
	return &Storage{db: tx}
}

func (s *Storage) Insert(ctx context.Context, sessionKey string, uid uuid.UUID) error{
	sql := "INSERT INTO auth.sessions (session_key,uid) VALUES ($1,$2);"
	_, err := s.db.Exec(ctx, sql, sessionKey,uid)
	if err != nil{
		slog.Error(tag("Storage sessions error: %v",err))
	}
	return err
}

func (s *Storage) Delete(ctx context.Context, id int) error{
	sql := "DELETE * FROM auth.sessions WHERE id=$1;"
	_, err := s.db.Exec(ctx, sql, id)
	if err != nil{
		slog.Error(tag("Storage sessions error: %v",err))
	}
	return err
}

func (s *Storage) Select(ctx context.Context, id int) (*models.Session,error){
	sql := "SELECT * FROM auth.sessions WHERE id=$1;"
	row := s.db.QueryRow(ctx, sql, id)
	res, err := scanner.Row(row,models.SessionFactory)
	if err != nil{
		slog.Error(tag("Storage sessions error: %v",err))
		return nil, err
	}
	return res, nil
}