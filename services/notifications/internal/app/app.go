// Package app configures and runs application.
package app

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/modulix-systems/goose-talk/internal/config"
	rmqController "github.com/modulix-systems/goose-talk/internal/controller/rmq"
	mailclient "github.com/modulix-systems/goose-talk/internal/gateways/mail"
	"github.com/modulix-systems/goose-talk/internal/services/mail"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/modulix-systems/goose-talk/rabbitmq"
)

func Run(cfg *config.Config) {
	log := logger.New(cfg.Log.Level)
	rmq, err := rabbitmq.New(cfg.RabbitMQ.Url)
	if err != nil {
		log.Fatal("app - New - rabbitmq.New: rabbitmq startup failed", "err", err)
	}
	defer rmq.Close()

	mailClient := mailclient.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.App.Name, cfg.App.Url)
	mailService := mail.New(mailClient, log)

	rmqServer := rabbitmq.NewServer(rmq)
	rmqController.Register(rmqServer, mailService, log)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)

	select {
	case s := <-interrupt:
		log.Info("app - Run - interrupt signal", "signalName", s.String())
	case err = <-rmqServer.ServeErr:
		log.Error("app - Run - rmqServer.ServeErr", "err", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rmqServer.Stop(ctx)
}
