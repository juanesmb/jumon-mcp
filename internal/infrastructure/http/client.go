package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	stdhttp "net/http"
	"time"
)

type Config struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	MaxRetryDelay time.Duration
	UserAgent     string
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:       30 * time.Second,
		MaxRetries:    2,
		RetryDelay:    time.Second,
		MaxRetryDelay: 10 * time.Second,
		UserAgent:     "jumon-mcp/1.0",
	}
}

type Client struct {
	config *Config
	client *stdhttp.Client
}

func NewClient(config *Config, transport stdhttp.RoundTripper) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	rt := transport
	if rt == nil {
		rt = stdhttp.DefaultTransport
	}
	return &Client{
		config: config,
		client: &stdhttp.Client{
			Timeout:   config.Timeout,
			Transport: rt,
		},
	}
}

func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.do(ctx, stdhttp.MethodGet, url, nil, headers)
}

func (c *Client) Post(ctx context.Context, url string, body any, headers map[string]string) (*Response, error) {
	return c.do(ctx, stdhttp.MethodPost, url, body, headers)
}

func (c *Client) do(ctx context.Context, method, url string, body any, headers map[string]string) (*Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		var payload io.Reader
		if body != nil {
			raw, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body: %w", err)
			}
			payload = bytes.NewReader(raw)
		}

		req, err := stdhttp.NewRequestWithContext(ctx, method, url, payload)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.config.UserAgent)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.client.Do(req.WithContext(ContextWithAttempt(ctx, attempt+1)))
		if err != nil {
			lastErr = fmt.Errorf("perform request: %w", err)
			if attempt < c.config.MaxRetries {
				c.wait(attempt)
				continue
			}
			return nil, lastErr
		}

		rawBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			if attempt < c.config.MaxRetries {
				c.wait(attempt)
				continue
			}
			return nil, lastErr
		}

		if shouldRetry(resp.StatusCode) && attempt < c.config.MaxRetries {
			lastErr = fmt.Errorf("received retryable status %d", resp.StatusCode)
			c.wait(attempt)
			continue
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       rawBody,
		}, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func shouldRetry(statusCode int) bool {
	return statusCode >= 500 || statusCode == 429 || statusCode == 408
}

func (c *Client) wait(attempt int) {
	delay := time.Duration(float64(c.config.RetryDelay) * math.Pow(2, float64(attempt)))
	if delay > c.config.MaxRetryDelay {
		delay = c.config.MaxRetryDelay
	}
	time.Sleep(delay)
}
