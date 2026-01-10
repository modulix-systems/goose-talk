package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseUrl         string
	bearerAuthToken string
	baseClient      *http.Client
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

func (c *Client) makeRequest(path string, method string, dst any, query url.Values, body any) error {
	method = strings.ToUpper(method)
	reqUrl, err := url.JoinPath(c.baseUrl, path)
	if err != nil {
		return fmt.Errorf("httpclient - %s '%s' - url.JoinPath: %w", method, path, err)
	}
	reqUrl += fmt.Sprintf("?%s", query.Encode())

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return fmt.Errorf("httpclient - %s '%s' - http.NewRequest: %w", method, path, err)
	}

	if c.bearerAuthToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.bearerAuthToken)
	}

	if method == "POST" {
		if body == nil {
			body = map[string]any{}
		}
		encodedBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("httpclient - %s '%s' - json.Marshal: %w", method, path, err)
		}
		bufferedBody := bytes.NewBuffer(encodedBody)
		req.Body = io.NopCloser(bufferedBody)
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.baseClient.Do(req)
	if err != nil {
		return fmt.Errorf("httpclient - %s '%s' - baseClient.Do: %w", method, path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("httpclient - %s '%s' - request failed with status '%d'", method, path, resp.StatusCode)
	}

	if dst == nil {
		return nil
	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("httpclient - %s '%s' - %s - io.ReadAll: %w", method, path, resp.Status, err)
	}

	if err := json.Unmarshal(buffer, &dst); err != nil {
		return fmt.Errorf("httpclient - %s '%s' - %s - json.Unmarshal('%s'): %w", method, path, resp.Status, string(buffer), err)
	}

	return nil
}

func (c *Client) Get(path string, query url.Values, dst any) error {
	return c.makeRequest(path, "GET", dst, query, nil)
}

func (c *Client) Post(path string, data any, dst any) error {
	return c.makeRequest(path, "POST", dst, url.Values{}, data)
}
