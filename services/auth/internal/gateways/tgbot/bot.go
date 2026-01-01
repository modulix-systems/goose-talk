package tgbot

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

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

	var response GetMeResponse
	err := httpClient.Get("getMe", url.Values{}, &response)
	if err != nil {
		return nil, fmt.Errorf("tgbot - New - getMe: %w", err)
	}

	return &Client{
		httpClient: httpClient,
		botUrl:     "https://t.me/" + response.Result.Username,
	}, nil
}

func (c *Client) GetStartLinkWithCode(code string) string {
	return c.botUrl + "/?start=" + code
}

func (c *Client) SendTextMsg(ctx context.Context, chatId string, text string) error {
	err := c.httpClient.Post("sendMessage", map[string]string{"chat_id": chatId, "text": text}, nil)
	if err != nil {
		return fmt.Errorf("tgbot - SendTextMsg - sendMessage: %w", err)
	}
	return nil
}

func (c *Client) GetLatestMsg(ctx context.Context) (*gateways.TelegramMsg, error) {
	query := url.Values{}
	query.Add("allowed_updates", `["message"]`)
	query.Add("offset", "-1")

	var response GetUpdatesResponse
	err := c.httpClient.Get("getUpdates", query, &response)
	if err != nil {
		return nil, fmt.Errorf("tgbot - GetLatestMsg - getUpdates: %w", err)
	}

	message := response.Result[0].Message
	return &gateways.TelegramMsg{
		DateSent: time.Unix(int64(message.Date), 0),
		Text:     message.Text,
		ChatId:   strconv.Itoa(message.Chat.ID),
	}, nil
}
