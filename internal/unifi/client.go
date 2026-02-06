package unifi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ilmax/unifi-client-go/pkg/config"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/ilmax/unifi-client-go/pkg/sitemanager"
)

// Config configures the UniFi API client.
type Config struct {
	APIKey        string
	BaseURL       string
	UserAgent     string
	Timeout       time.Duration
	AllowInsecure bool
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
	if cfg.AllowInsecure {
		opts = append(opts, config.ConfigHTTPClient(insecureHTTPClient(cfg.Timeout)))
	} else if cfg.Timeout != 0 {
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

func insecureHTTPClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = config.DefaultTimeout
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: timeout}
	}
	cloned := transport.Clone()
	if cloned.TLSClientConfig == nil {
		cloned.TLSClientConfig = &tls.Config{}
	}
	cloned.TLSClientConfig.InsecureSkipVerify = true
	return &http.Client{
		Timeout:   timeout,
		Transport: cloned,
	}
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
	traceEnabled := isTraceEnabled()
	tflog.Debug(ctx, "UniFi API request", map[string]any{
		"method": method,
		"path":   path,
	})

	var bodyReader io.Reader
	var requestBody []byte
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = jsonBody
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

	if traceEnabled {
		if len(requestBody) > 0 {
			tflog.Trace(ctx, "UniFi API request body", map[string]any{
				"method": method,
				"path":   path,
				"body":   prettyJSONBytes(requestBody),
			})
		} else {
			tflog.Trace(ctx, "UniFi API request body", map[string]any{
				"method": method,
				"path":   path,
				"body":   "",
			})
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	requestID := resp.Header.Get("X-Request-Id")

	var respBody []byte
	shouldReadBody := traceEnabled || result != nil || resp.StatusCode >= 400
	if shouldReadBody {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	tflog.Debug(ctx, "UniFi API response", map[string]any{
		"method":       method,
		"path":         path,
		"status":       resp.StatusCode,
		"content_type": contentType,
		"request_id":   requestID,
		"body_bytes":   len(respBody),
	})

	if traceEnabled {
		if len(respBody) > 0 {
			tflog.Trace(ctx, "UniFi API response body", map[string]any{
				"method": method,
				"path":   path,
				"body":   prettyJSONBytes(respBody),
			})
		} else {
			tflog.Trace(ctx, "UniFi API response body", map[string]any{
				"method": method,
				"path":   path,
				"body":   "",
			})
		}
	}

	if strings.Contains(strings.ToLower(contentType), "text/html") {
		tflog.Warn(ctx, "UniFi API returned HTML response", map[string]any{
			"method":       method,
			"path":         path,
			"status":       resp.StatusCode,
			"content_type": contentType,
			"request_id":   requestID,
		})
		if resp.StatusCode < 400 {
			return fmt.Errorf("unifi api returned HTML response (content-type %q). check api_url or proxy", contentType)
		}
	}

	if resp.StatusCode >= 400 {
		return errors.NewAPIError(
			resp.StatusCode,
			string(respBody),
			requestID,
		)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func prettyJSONBytes(input []byte) string {
	var out bytes.Buffer
	if err := json.Indent(&out, input, "", "  "); err != nil {
		return string(input)
	}
	return out.String()
}

func isTraceEnabled() bool {
	level := strings.TrimSpace(os.Getenv("TF_LOG_PROVIDER"))
	if level == "" {
		level = strings.TrimSpace(os.Getenv("TF_LOG"))
	}
	return strings.EqualFold(level, "TRACE")
}
