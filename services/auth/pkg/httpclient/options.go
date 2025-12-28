package httpclient

import "net/http"

type Option func(c *Client)

func BaseClient(client *http.Client) Option {
	return func(c *Client) {
		c.baseClient = client
	}
}