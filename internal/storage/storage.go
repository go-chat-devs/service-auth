package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chat-devs/service-auth/internal/storage/sessions"
	twofactortotp "github.com/go-chat-devs/service-auth/internal/storage/two_factor_totp"
	"github.com/go-chat-devs/service-auth/internal/storage/users"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	gw   *gateway
	pool *pgxpool.Pool

	users         *users.Storage
	sessions      *sessions.Storage
	twoFactorTotp *twofactortotp.Storage
}

func New(ctx context.Context) (*Storage, error) {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		slog.Error(fmt.Sprintf("error parsing connection config: %v", err))
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("error creating connection pool: %v", err))
		return nil, err
	}

	return &Storage{
		gw:            newGateway(ctx),
		pool:          pool,
		users:         users.New(pool),
		sessions:      sessions.New(pool),
		twoFactorTotp: twofactortotp.New(pool),
	}, nil
}

func (s *Storage) CreateUser(ctx context.Context, email, password string) error
func (s *Storage) AuthenticateUser(ctx context.Context, email, password string) (string, error)
func (s *Storage) ValidateTOTP(ctx context.Context, tempToken string, code string) (string, error)
func (s *Storage) GetUserUID(ctx context.Context, token string) (uuid.UUID, error)
func (s *Storage) DeleteUser(ctx context.Context, token string, password string) error
