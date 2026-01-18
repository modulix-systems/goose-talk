package mail

import (
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/logger"
)

type Service struct {
	mailClient gateways.MailClient
	log        logger.Interface
}

func New(mailClient gateways.MailClient, log logger.Interface) *Service {
	return &Service{
		mailClient: mailClient,
		log:        log,
	}
}
