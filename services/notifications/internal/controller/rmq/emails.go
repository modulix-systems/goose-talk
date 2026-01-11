package rmq

import (
	"encoding/json"

	notificationsContracts "github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
	"github.com/modulix-systems/goose-talk/internal/services/mail"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/rabbitmq/amqp091-go"
)

type EmailsHandler struct {
	mailService *mail.Service
	log         logger.Interface
}

func NewEmailsHandler(mailService *mail.Service, log logger.Interface) *EmailsHandler {
	return &EmailsHandler{mailService, log}
}

func (c *EmailsHandler) Handle(delivery amqp091.Delivery) {
	var email notificationsContracts.EmailMessage
	if err := json.Unmarshal(delivery.Body, &email); err != nil {
		c.log.Error("rmq - EmailHandler.Handle - parse delivery", "err", err)
		return
	}
	if err := c.mailService.SendMail(email.Name, email.Data, email.To); err != nil {
		c.log.Error("rmq - EmailHandler.Handle - SendMail", "err", err)
	}
}
