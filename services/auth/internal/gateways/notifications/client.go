package notifications

import (
	"context"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/logger"
)

type Client struct {
	log logger.Interface
}

func New(log logger.Interface) *Client {
	return &Client{log}
}

func (c *Client) SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error {
	c.log.Info("Sending sign up confirmation email", "to", to)
	return nil
}

func (c *Client) SendGreetingEmail(ctx context.Context, to string, name string) error {
	c.log.Info("Sending greeting email", "to", to, "name", name)
	return nil
}

func (c *Client) Send2FAEmail(ctx context.Context, to string, otp string) error {
	c.log.Info("Sending 2FA email", "to", to)
	return nil
}

func (c *Client) SendAccDeactivationEmail(ctx context.Context, to string) error {
	c.log.Info("Sending account deactivation email", "to", to)
	return nil
}

func (c *Client) SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.AuthSession) error {
	c.log.Info("Sending new device sign in email", "to", to, "sessionID", newSession.Id)
	return nil
}
