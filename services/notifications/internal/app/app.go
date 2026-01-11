// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/logger"
	rabbitmq "github.com/modulix-systems/goose-talk/pkg/rmq"
	"github.com/modulix-systems/goose-talk/postgres"
)

func Run(cfg *config.Config) {
	log := logger.New(cfg.Log.Level)
	rmq, err := rabbitmq.New(cfg.RabbitMQ.Url)
	if err != nil {
		log.Fatal("app - New - rabbitmq.New: rabbitmq startup failed", "err", err)
	}
}
