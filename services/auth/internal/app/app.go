// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/grpcserver"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	pg, err := postgres.New(cfg.PG.Dsn, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()
	authService := auth.New()

	grpcServer := grpcserver.New(l, cfg.Server.Port)
	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-grpcServer.ServeErr:
		l.Error(fmt.Errorf("app - Run - grpcServer.ServeErr: %w", err))
	}

	// Shutdown
	grpcServer.Stop()
}
