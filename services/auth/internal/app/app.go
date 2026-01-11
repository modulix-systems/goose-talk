// Package app configures and runs application.
package app

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/modulix-systems/goose-talk/internal/config"
	rpc_v1 "github.com/modulix-systems/goose-talk/internal/controller/grpc/v1"
	"github.com/modulix-systems/goose-talk/internal/gateways/geoip"
	"github.com/modulix-systems/goose-talk/internal/gateways/notifications"
	"github.com/modulix-systems/goose-talk/internal/gateways/security"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/pgrepos"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/redisrepos"
	"github.com/modulix-systems/goose-talk/internal/gateways/tgbot"
	"github.com/modulix-systems/goose-talk/internal/gateways/webauthn"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/modulix-systems/goose-talk/pkg/grpcserver"
	"github.com/modulix-systems/goose-talk/pkg/redis"
	"github.com/modulix-systems/goose-talk/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	log := logger.New(cfg.Log.Level)

	pgOpts := []postgres.Option{}
	if cfg.Postgres.MaxPoolSize != 0 {
		pgOpts = append(pgOpts, postgres.MaxPoolSize(cfg.Postgres.MaxPoolSize))
	}
	pg, err := postgres.New(cfg.Postgres.Url, pgOpts...)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	redisOpts := []redis.Option{}
	if cfg.Redis.MaxPoolSize != 0 {
		redisOpts = append(redisOpts, redis.MaxPoolSize(cfg.Redis.MaxPoolSize))
	}
	rdb, err := redis.New(cfg.Redis.Url, redisOpts...)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - redis.New: %w", err))
	}
	defer rdb.Close()

	pgRepos := pgrepos.New(pg)
	redisRepos := redisrepos.New(rdb)

	appUrl, err := url.Parse(cfg.App.Url)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - url.Parse: %w", err))
	}

	notificationsClient := notifications.New(log)
	geoipClient := geoip.New()
	securityProvider := security.New(cfg.TotpTTL, config.OTP_LENGTH, cfg.App.Name)
	webauthnProvider := webauthn.New(cfg.App.Name, appUrl.Host, []string{appUrl.Host})

	tgBotClient, err := tgbot.New(cfg.Tgbot.Token)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - tgbot.New: %w", err))
	}

	authService := auth.New(
		pgRepos.Users,
		redisRepos.AuthSessions,
		redisRepos.QRLoginTokens,
		redisRepos.Otp,
		redisRepos.PasskeySession,
		notificationsClient,
		webauthnProvider,
		securityProvider,
		tgBotClient,
		geoipClient,

		cfg.OtpTTL,
		cfg.LoginTokenTTL,
		cfg.DefaultSessionTTL,
		cfg.LongLivedSessionTTL,
		log,
	)

	validate := validator.New(validator.WithRequiredStructEnabled())
	grpcServer := grpcserver.New(log, cfg.Port)
	rpc_v1.Register(grpcServer, authService, log, validate)

	go grpcServer.Run()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - interrupt signal", "signalName", s.String())
	case err = <-grpcServer.ServeErr:
		log.Error(fmt.Errorf("app - Run - grpcServer.ServeErr: %w", err))
	}

	// Shutdown
	grpcServer.Stop()
}
