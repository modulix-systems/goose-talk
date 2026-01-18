package rmq

import (
	notificationsContracts "github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
	"github.com/modulix-systems/goose-talk/internal/services/mail"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/modulix-systems/goose-talk/rabbitmq"
)

func Register(
	server *rabbitmq.Server,
	mailService *mail.Service,
	log logger.Interface,
) {
	contracts := notificationsContracts.New()
	emailsHandler := NewEmailsHandler(mailService, log)
	server.RegisterQueue(contracts.Queues.Emails, emailsHandler)
}
