package sendgrid

import (
	"github.com/modulix-systems/goose-talk/httpclient"
)

type Client struct {
	httpClient *httpclient.Client
}

func New(apiKey string) *Client {
	httpClient := httpclient.New("https://api.sendgrid.com/v3/", httpclient.BearerAuth(apiKey))
	return &Client{httpClient}
}
