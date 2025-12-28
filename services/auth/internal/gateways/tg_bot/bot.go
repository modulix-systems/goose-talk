package tg_bot

import (
	"context"
	"fmt"

	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/pkg/httpclient"
)

type Client struct {
	botUrl     string
	httpClient *httpclient.Client
}

func New(botToken string) (*Client, error) {
	if botToken == "" {
		panic("botToken is not provided")
	}

	httpClient := httpclient.New("https://api.telegram.org/bot" + botToken)

	resp, err := httpClient.Get("/getMe", map[string][]string{})
	if err != nil {
		return nil, fmt.Errorf("tg_bot - New - getMe: %w", err)
	}

	botUsername := resp["result"].(map[string]any)["username"].(string)

	return &Client{
		httpClient: httpClient,
		botUrl:     "https://t.me/" + botUsername,
	}, nil
}

func (c *Client) GetStartLinkWithCode(code string) string {
	return c.botUrl + "/?start=" + code
}

func (c *Client) SendTextMsg(ctx context.Context, chatId string, text string) error {
	return nil
}

func (c *Client) GetLatestMsg(ctx context.Context) (*gateways.TelegramMsg, error) {
	return nil, nil
}
