package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	notificationsContracts "github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/modulix-systems/goose-talk/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	channel   *amqp091.Channel
	contracts *notificationsContracts.Contracts
	log       logger.Interface
}

func New(rmq *rabbitmq.RabbitMQ, log logger.Interface) (*Client, error) {
	channel, err := rmq.NewChannel()
	if err != nil {
		return nil, fmt.Errorf("notifications.New - rmq.NewChannel: %w", err)
	}

	contracts := notificationsContracts.New()

	if _, err := rmq.QueueDeclare(contracts.Queues.Emails, channel); err != nil {
		return nil, fmt.Errorf("notifications.New - rmq.QueueDeclare - declare emails queue: %w", err)
	}

	if _, err := rmq.QueueDeclare(contracts.Queues.Notifications, channel); err != nil {
		return nil, fmt.Errorf("notifications.New - rmq.QueueDeclare - declare notifications queue: %w", err)
	}

	return &Client{channel: channel, contracts: contracts, log: log}, nil
}

func (c *Client) sendEmailNotice(ctx context.Context, typ notificationsContracts.EmailType, to string, payload any, lang string) error {
	c.log.Info("Sending email notice", "to", to, "typ", typ)
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifications.Client.sendEmailNotice - %s - marshal payload: %w", typ, err)
	}
	message := notificationsContracts.EmailMessage{
		Language: notificationsContracts.Language(lang),
		Type:     typ,
		Data:     payloadJson,
		To:       to,
	}
	messageJson, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("notifications.Client.sendEmailNotice - %s - marshal email message: %w", typ, err)
	}

	publishing := amqp091.Publishing{Body: messageJson}
	if err = c.channel.PublishWithContext(ctx, "", c.contracts.Queues.Emails.Name, false, false, publishing); err != nil {
		return fmt.Errorf("notifications.Client.sendEmailNotice - %s - publish to queue: %w", typ, err)
	}
	c.log.Info("Email notice published", "to", to, "typ", typ)

	return nil
}

func (c *Client) SendEmailVerifyEmail(ctx context.Context, to, otp string) error {
	payload := notificationsContracts.EmailVerifyNotice{
		Code: otp,
	}

	return c.sendEmailNotice(
		ctx,
		notificationsContracts.EMAIL_TYPE_VERIFY_EMAIL,
		to,
		payload,
		"",
	)
}

func (c *Client) SendSignUpEmail(
	ctx context.Context,
	user *entity.User,
) error {
	payload := notificationsContracts.SignUpNotice{
		Username: user.Username,
	}

	return c.sendEmailNotice(
		ctx,
		notificationsContracts.EMAIL_TYPE_SIGN_UP,
		user.Email,
		payload,
		user.Language,
	)
}

func (c *Client) SendConfirmEmailTwoFaEmail(
	ctx context.Context,
	to, username, otp, lang string,
) error {
	payload := notificationsContracts.EmailTwoFaNotice{
		Username: username,
		Code:     otp,
	}

	return c.sendEmailNotice(
		ctx,
		notificationsContracts.EMAIL_TYPE_EMAIL_TWO_FA,
		to,
		payload,
		lang,
	)
}

func (c *Client) SendAccountDeactivatedEmail(
	ctx context.Context,
	to, username, lang string,
) error {
	payload := notificationsContracts.AccountDeactivatedNotice{
		Username: username,
	}

	return c.sendEmailNotice(
		ctx,
		notificationsContracts.EMAIL_TYPE_ACCOUNT_DEACTIVATED,
		to,
		payload,
		lang,
	)
}

func (c *Client) SendLoginNewDeviceEmail(
	ctx context.Context,
	to, username string,
	newSession *entity.AuthSession,
	lang string,
) error {
	payload := notificationsContracts.LoginNewDeviceNotice{
		Username:   username,
		IpAddr:     newSession.IpAddr,
		DeviceInfo: newSession.DeviceInfo,
		Location:   newSession.Location,
	}

	return c.sendEmailNotice(
		ctx,
		notificationsContracts.EMAIL_TYPE_LOGIN_NEW_DEVICE,
		to,
		payload,
		lang,
	)
}
