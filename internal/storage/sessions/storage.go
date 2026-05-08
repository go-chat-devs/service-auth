package sessions

import (
	"context"
	"log/slog"

	"github.com/go-chat-devs/service-auth/internal/models"
	"github.com/go-chat-devs/service-auth/internal/scanner"
	"github.com/go-chat-devs/service-auth/internal/storage/db"
	"github.com/go-chat-devs/service-auth/internal/tagger"
	"github.com/go-chat-devs/service-auth/internal/token"
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

func (s *Storage) Insert(ctx context.Context, userID int) (*models.Session, error) {
	sessionKey := token.Generate()
	sessionKeySlice := sessionKey[:] 
	sql := "INSERT INTO auth.sessions(session_key, user_id) VALUES($1, $2)"
	_, err := s.db.Exec(ctx, sql, sessionKeySlice, userID)
	if err != nil {
		slog.Error(tag("insert error: %v", err))
	}
	return &models.Session{
		UserID:     userID,
		SessionKey: sessionKey,
	}, err
}

func (s *Storage) Delete(ctx context.Context, sessionKey token.Token) error {
	sessionKeySlice := sessionKey[:]
	sql := "DELETE FROM auth.sessions WHERE session_key=$1"
	_, err := s.db.Exec(ctx, sql, sessionKeySlice)
	if err != nil {
		slog.Error(tag("delete error: %v", err))
	}
	return err
}

func (s *Storage) Select(ctx context.Context, sessionKey token.Token) (*models.Session, error) {
	sql := "SELECT * FROM auth.sessions WHERE sessionKey=$1"
	row := s.db.QueryRow(ctx, sql, sessionKey)
	res, err := scanner.Row(row, models.SessionFactory)
	if err != nil {
		slog.Error(tag("select error: %v", err))
		return nil, err
	}
	return res, nil
}


func (s *Storage) DeleteAll(ctx context.Context, userID int) error {
	sql := "DELETE FROM auth.sessions WHERE user_id=$1"
	_,err := s.db.Exec(ctx,sql,userID)
	if err != nil{
		slog.Error(tag("delete error: %v",err))
	}
	return err
}