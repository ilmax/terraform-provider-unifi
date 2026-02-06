package unifi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ilmax/unifi-client-go/pkg/config"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/ilmax/unifi-client-go/pkg/sitemanager"
)

// Config configures the UniFi API client.
type Config struct {
	APIKey    string
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
}

// Client provides a minimal API client for the UniFi cloud endpoints.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

// NewClient creates a new UniFi API client.
func NewClient(cfg Config) (*Client, error) {
	sdkCfg := config.New()
	opts := []config.ConfigOption{
		config.ConfigAPIKey(cfg.APIKey),
		config.ConfigBaseURL(cfg.BaseURL),
	}
	if cfg.UserAgent != "" {
		opts = append(opts, config.ConfigUserAgent(cfg.UserAgent))
	}
	if cfg.Timeout != 0 {
		opts = append(opts, config.ConfigTimeout(cfg.Timeout))
	}
	if err := sdkCfg.Init(opts); err != nil {
		return nil, err
	}
	if sdkCfg.APIKey == "" {
		return nil, errors.ErrEmptyAPIKey
	}
	if sdkCfg.BaseURL == "" {
		sdkCfg.BaseURL = sitemanager.DefaultBaseURL
	}

	return &Client{
		httpClient: sdkCfg.HTTPClient,
		baseURL:    strings.TrimSuffix(sdkCfg.BaseURL, "/"),
		apiKey:     sdkCfg.APIKey,
		userAgent:  sdkCfg.UserAgent,
	}, nil
}

// Get sends a GET request.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// Post sends a POST request.
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Put sends a PUT request.
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodDelete, path, nil, result)
}

func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return errors.NewAPIError(
			resp.StatusCode,
			string(respBody),
			resp.Header.Get("X-Request-Id"),
		)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
