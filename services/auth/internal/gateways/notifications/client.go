package notifications

import (
	"context"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/logger"
)

type Client struct {
	log logger.Interface
}

func New(log logger.Interface) *Client {
	return &Client{log}
}

func (c *Client) SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error {
	c.log.Info("Sending sign up confirmation email", "to", to, "otp", otp)
	return nil
}

func (c *Client) SendGreetingEmail(ctx context.Context, to string, name string) error {
	return nil
}

func (c *Client) Send2FAEmail(ctx context.Context, to string, otp string) error {
	return nil
}

func (c *Client) SendAccDeactivationEmail(ctx context.Context, to string) error {
	return nil
}

func (c *Client) SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.AuthSession) error {
	return nil
}
