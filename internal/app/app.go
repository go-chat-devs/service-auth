package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/go-chat-devs/service-auth/internal/app/grpc"
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
	//Implement this in future
	return &App{}
}
