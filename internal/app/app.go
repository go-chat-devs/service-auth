package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/go-chat-devs/service-auth/internal/app/grpc"
	"github.com/go-chat-devs/service-auth/internal/services/auth"
	"github.com/go-chat-devs/service-auth/internal/storage"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	ctx context.Context,
	log *slog.Logger,
	grpcPort int,
	storageURL string,
	tokenTTL time.Duration,
) *App {
	storage, err := storage.New(ctx)
	if err != nil {
		panic(err)
	}

	authserivce := auth.New(log, storage, storage, storage,tokenTTL)
	grpcApp := grpcapp.New(log, authserivce, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
