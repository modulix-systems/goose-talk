package mailer

import "github.com/modulix-systems/goose-talk/logger"

type MailService struct {
	log logger.Interface
}

func New() *MailService {
	return &MailService{}
}
