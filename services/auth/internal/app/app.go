// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modulix-systems/goose-talk/internal/config"
	rpc_v1 "github.com/modulix-systems/goose-talk/internal/controller/grpc/v1"
	postgres_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/grpcserver"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	log := logger.New(cfg.Log.Level)

	pg, err := postgres.New(cfg.PG.Dsn, postgres.MaxPoolSize(cfg.PG.MaxPoolSize))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()
	repositories := postgres_repos.New(pg)
	authService := auth.New(
		repositories.Users,
	)
	grpcServer := grpcserver.New(log, cfg.Server.Port)
	rpc_v1.Register(grpcServer, authService, log)
	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: %s", s.String())
	case err = <-grpcServer.ServeErr:
		log.Error(fmt.Errorf("app - Run - grpcServer.ServeErr: %w", err))
	}

	// Shutdown
	grpcServer.Stop()
}
