package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseUrl    string
	baseClient *http.Client
}

func New(baseUrl string, opts ...Option) *Client {
	client := &Client{baseUrl: baseUrl}

	for _, opt := range opts {
		opt(client)
	}

	if client.baseClient == nil {
		client.baseClient = http.DefaultClient
	}

	return client
}

func (c *Client) Get(path string, query url.Values) (map[string]any, error) {
	reqUrl, err := url.JoinPath(c.baseUrl, path, query.Encode())
	if err != nil {
		return nil, fmt.Errorf("httpclient - Get - url.JoinPath: %w", err)
	}
	resp, err := c.baseClient.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("httpclient - Get '%s' - baseClient.Get: %w", reqUrl, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("httpclient - Get '%s' - request failed with status: %d", reqUrl, resp.StatusCode)
	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpclient - Get '%s' - %s - io.ReadAll: %w", reqUrl, resp.Status, err)
	}

	var data map[string]any
	if err := json.Unmarshal(buffer, &data); err != nil {
		return nil, fmt.Errorf("httpclient - Get '%s' - %s - json.Unmarshal('%s'): %w", reqUrl, resp.Status, string(buffer), err)
	}

	return data, nil
}
