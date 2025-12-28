// Package app configures and runs application.
package app

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/modulix-systems/goose-talk/internal/config"
	rpc_v1 "github.com/modulix-systems/goose-talk/internal/controller/grpc/v1"
	geoip "github.com/modulix-systems/goose-talk/internal/gateways/geo-ip"
	"github.com/modulix-systems/goose-talk/internal/gateways/notifications"
	"github.com/modulix-systems/goose-talk/internal/gateways/security"
	postgres_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres"
	redis_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/redis"
	"github.com/modulix-systems/goose-talk/internal/gateways/tg_bot"
	"github.com/modulix-systems/goose-talk/internal/gateways/webauthn"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/grpcserver"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
	"github.com/modulix-systems/goose-talk/pkg/redis"
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

	pgRepos := postgres_repos.New(pg)
	redisRepos := redis_repos.New(rdb)

	appUrl, err := url.Parse(cfg.App.Url)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - url.Parse: %w", err))
	}

	notificationsClient := notifications.New()
	geoipClient := geoip.New()
	securityProvider := security.New(cfg.TotpTTL, config.OTP_LENGTH)
	webauthnProvider := webauthn.New(cfg.App.Name, appUrl.Host, []string{appUrl.Host})

	tgBotClient, err := tg_bot.New(cfg.Tgbot.Token)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - tg_bot.New: %w", err))
	}

	authService := auth.New(
		pgRepos.Users,
		redisRepos.Sessions,
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

	grpcServer := grpcserver.New(log, cfg.Port)
	rpc_v1.Register(grpcServer, authService, log)
	
	go grpcServer.Run()

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
