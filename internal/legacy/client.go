// Package legacy is a deliberately tiny client for the controller's legacy
// (non-Integrations) API under /proxy/network. It exists only for settings the
// official Integrations API does not expose yet; delete each helper once the
// official API covers it. Authentication is the same X-API-Key header.
package legacy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// BasePath is where the legacy API lives on a UniFi OS console.
const BasePath = "/proxy/network"

// ConsoleURL reduces any console, legacy or Integrations API URL to the
// console base URL, so the same api_url works for both providers.
func ConsoleURL(apiURL string) string {
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/integrations")
	apiURL = strings.TrimSuffix(apiURL, "/integration")
	return strings.TrimSuffix(apiURL, BasePath)
}

// Client wraps the shared HTTP client with legacy-API conventions: the {meta,data}
// envelope and sites addressed by name instead of UUID.
type Client struct {
	api *unifi.Client

	mu    sync.Mutex
	sites map[string]string // Integrations site UUID -> legacy site name
}

// NewClient builds a Client for the console at apiURL.
func NewClient(cfg unifi.Config) (*Client, error) {
	cfg.BaseURL = ConsoleURL(cfg.BaseURL) + BasePath
	api, err := unifi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{api: api, sites: map[string]string{}}, nil
}

type envelope struct {
	Meta struct {
		RC  string `json:"rc"`
		Msg string `json:"msg"`
	} `json:"meta"`
	Data json.RawMessage `json:"data"`
}

func (c *Client) call(ctx context.Context, method, path string, body, out any) error {
	var env envelope
	var err error
	switch method {
	case "GET":
		err = c.api.Get(ctx, path, &env)
	case "PUT":
		err = c.api.Put(ctx, path, body, &env)
	case "POST":
		err = c.api.Post(ctx, path, body, &env)
	default:
		return fmt.Errorf("unsupported method %s", method)
	}
	if err != nil {
		return err
	}
	if env.Meta.RC != "ok" {
		return fmt.Errorf("legacy API %s %s: rc=%q msg=%q", method, path, env.Meta.RC, env.Meta.Msg)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// SiteName resolves an Integrations API site UUID (the site_id used across this
// repo) to the name the legacy API puts in its URLs, e.g. "default".
func (c *Client) SiteName(ctx context.Context, siteID string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if name, ok := c.sites[siteID]; ok {
		return name, nil
	}
	var sites []struct {
		ExternalID string `json:"external_id"`
		Name       string `json:"name"`
	}
	if err := c.call(ctx, "GET", "/api/self/sites", nil, &sites); err != nil {
		return "", err
	}
	for _, s := range sites {
		c.sites[s.ExternalID] = s.Name
	}
	if name, ok := c.sites[siteID]; ok {
		return name, nil
	}
	return "", fmt.Errorf("no site with id %q on this console", siteID)
}

// GetSetting reads the per-site setting document with the given key (e.g.
// "teleport") from rest/setting into out.
func (c *Client) GetSetting(ctx context.Context, siteID, key string, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	var all []json.RawMessage
	if err := c.call(ctx, "GET", fmt.Sprintf("/api/s/%s/rest/setting", site), nil, &all); err != nil {
		return err
	}
	for _, raw := range all {
		var k struct {
			Key string `json:"key"`
		}
		if json.Unmarshal(raw, &k) == nil && k.Key == key {
			return json.Unmarshal(raw, out)
		}
	}
	return fmt.Errorf("setting %q not found on site %q", key, site)
}

// UpdateSetting PUTs fields into the setting document id and decodes the
// updated document into out.
func (c *Client) UpdateSetting(ctx context.Context, siteID, key, id string, fields, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.single(ctx, "PUT", fmt.Sprintf("/api/s/%s/rest/setting/%s/%s", site, key, id), fields, out)
}

// List reads every document of a rest/<collection> endpoint (e.g. "user").
func (c *Client) List(ctx context.Context, siteID, collection string, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.call(ctx, "GET", fmt.Sprintf("/api/s/%s/rest/%s", site, collection), nil, out)
}

// Create POSTs a document to rest/<collection> and decodes the stored document.
func (c *Client) Create(ctx context.Context, siteID, collection string, body, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.single(ctx, "POST", fmt.Sprintf("/api/s/%s/rest/%s", site, collection), body, out)
}

// Update PUTs fields into rest/<collection>/<id> and decodes the stored document.
func (c *Client) Update(ctx context.Context, siteID, collection, id string, body, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.single(ctx, "PUT", fmt.Sprintf("/api/s/%s/rest/%s/%s", site, collection, id), body, out)
}

// single runs a call whose response data holds exactly one document.
func (c *Client) single(ctx context.Context, method, path string, body, out any) error {
	var docs []json.RawMessage
	if err := c.call(ctx, method, path, body, &docs); err != nil {
		return err
	}
	if len(docs) != 1 {
		return fmt.Errorf("legacy API %s %s returned %d documents, want 1", method, path, len(docs))
	}
	return json.Unmarshal(docs[0], out)
}

// Command runs a manager command such as {"cmd":"forget-sta","macs":[...]} on
// cmd/<manager> (e.g. "stamgr"). Prefer typed helpers; this exists for cleanup.
func (c *Client) Command(ctx context.Context, siteID, manager string, body any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.call(ctx, "POST", fmt.Sprintf("/api/s/%s/cmd/%s", site, manager), body, nil)
}

// Stat reads a stat/<name> report (e.g. "device"); live state is only listed
// there, rest/device does not support GET.
func (c *Client) Stat(ctx context.Context, siteID, name string, out any) error {
	site, err := c.SiteName(ctx, siteID)
	if err != nil {
		return err
	}
	return c.call(ctx, "GET", fmt.Sprintf("/api/s/%s/stat/%s", site, name), nil, out)
}
