package geoip

import (
	"fmt"
	"net/url"

	"github.com/modulix-systems/goose-talk/httpclient"
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

	var response GetLocationResponse
	err := c.httpClient.Get(ip, query, &response)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s, %s", response.City, response.Country), nil
}
