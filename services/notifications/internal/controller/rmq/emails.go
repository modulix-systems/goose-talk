package rmq

import (
	"context"
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

func (c *EmailsHandler) Handle(delivery amqp091.Delivery) error {
	var email notificationsContracts.EmailMessage
	if err := json.Unmarshal(delivery.Body, &email); err != nil {
		c.log.Error("rmq - EmailHandler.Handle - error parse delivery", "err", err)
		return err
	}

	c.log.Info("rmq - EmailHandler.Handle - handling new email", "type", email.Type, "to", email.To)

	if err := c.mailService.SendMail(context.Background(), email); err != nil {
		c.log.Error("rmq - EmailHandler.Handle - error sending mail", "err", err)
		return err
	}

	return nil
}
