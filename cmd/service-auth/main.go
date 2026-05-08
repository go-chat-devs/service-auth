package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chat-devs/service-auth/internal/app"
	"github.com/go-chat-devs/service-auth/internal/config"
	"github.com/go-chat-devs/service-auth/internal/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	ctx := context.Background()
	log.Info(
		"starting application",
		slog.String("env", cfg.Env),
		slog.Any("cfg", cfg),
		slog.Int("port", cfg.GRPC.Port),
	)
	application := app.New(ctx, log, cfg.GRPC.Port, cfg.StorageUrl, cfg.TokenTTL)

	go application.GRPCSrv.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	s := <-stop
	log.Info("stopping application", slog.String("signal", s.String()))

	application.GRPCSrv.Stop()
	log.Info("application stopped")

}
