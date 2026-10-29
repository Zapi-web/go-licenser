package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zapi-web/go-licenser/internal/config"
	"github.com/zapi-web/go-licenser/internal/database"
	"github.com/zapi-web/go-licenser/internal/httpapi"
	"github.com/zapi-web/go-licenser/internal/kgs"
	"github.com/zapi-web/go-licenser/internal/logger"
	"github.com/zapi-web/go-licenser/internal/service"
	"github.com/zapi-web/go-licenser/internal/signer"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.New()

	if err != nil {
		slog.Error("failed to initialize config", "err", err)
		return 1
	}

	slog.SetDefault(logger.New(cfg.LogLevel))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ed25519Signer, err := signer.New(cfg.PrivateKey, cfg.Version)
	if err != nil {
		slog.Error("failed to create signer", "err", err)
		return 1
	}

	dbConfig := database.Config{
		MaxConns:          cfg.MaxConns,
		MinConns:          cfg.MinConns,
		MaxConnLifeTime:   cfg.MaxConnLifeTime,
		MaxConnIdleTime:   cfg.MaxConnIdleTime,
		HealthCheckPeriod: cfg.HealthCheckPeriod,
		ConnectionTimeout: cfg.ConnectionTimeout,
	}

	db, err := database.New(ctx, cfg.ConnString, dbConfig)
	if err != nil {
		slog.Error("failed to connnect to a database", "err", err)
		return 1
	}
	defer db.Close()

	keyGeneration := kgs.New(cfg.KeyGroups, cfg.GroupsLen)

	licenserService := service.New(db, ed25519Signer, keyGeneration)

	api := httpapi.New(licenserService, cfg.AdminToken)
	server := httpapi.NewServer(cfg.Port, api.Handler(), cfg.ReadTimeout, cfg.WriteTimeout)

	serverError := make(chan error, 1)
	go func() {
		serverError <- server.RunServer()
	}()

	select {
	case err := <-serverError:
		if err != nil {
			slog.Error("received an error from http server", "err", err)
			return 1
		}
	case <-ctx.Done():
		slog.Info("received a signal, starting graceful shutdown")

		srvShutdownCtx, srvCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer srvCancel()

		if err := server.Shutdown(srvShutdownCtx); err != nil {
			slog.Error("failed to shutdown server", "err", err)
		}
	}

	slog.Debug("Server stopped")
	return 0
}
