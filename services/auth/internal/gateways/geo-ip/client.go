package geoip

import (
	"fmt"
	"net/url"

	"github.com/modulix-systems/goose-talk/pkg/httpclient"
)

type Client struct {
	httpClient *httpclient.Client
}

func New() *Client {
	httpClient := httpclient.New("http://ip-api.com/json/")
	return &Client{httpClient}
}

func (c *Client) GetLocationByIP(ip string) (string, error) {
	query := url.Values{}
	query.Set("fields", "city,country")
	resp, err := c.httpClient.Get(ip, query)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s, %s", resp["city"].(string), resp["country"].(string)), nil
}
