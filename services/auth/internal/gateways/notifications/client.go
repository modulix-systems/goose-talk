package notifications

import (
	"context"

	"github.com/modulix-systems/goose-talk/internal/entity"
)

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (client *Client) SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error {
	return nil
}

func (client *Client) SendGreetingEmail(ctx context.Context, to string, name string) error {
	return nil
}

func (client *Client) Send2FAEmail(ctx context.Context, to string, otp string) error {
	return nil
}

func (client *Client) SendAccDeactivationEmail(ctx context.Context, to string) error {
	return nil
}

func (client *Client) SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.AuthSession) error {
	return nil
}
