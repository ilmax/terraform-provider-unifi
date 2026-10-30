package legacy

import "context"

// Teleport is the per-site "teleport" setting. Field names and the subnet
// pattern come from the controller's own api/fields/Setting.json.
type Teleport struct {
	ID         string `json:"_id,omitempty"`
	Enabled    bool   `json:"enabled"`
	SubnetCIDR string `json:"subnet_cidr"`
}

func (c *Client) GetTeleport(ctx context.Context, siteID string) (Teleport, error) {
	var t Teleport
	err := c.GetSetting(ctx, siteID, "teleport", &t)
	return t, err
}

func (c *Client) UpdateTeleport(ctx context.Context, siteID, id string, t Teleport) (Teleport, error) {
	// An empty subnet_cidr must never be sent: the controller stores it and then
	// fails to parse it ("Could not parse []") when it lists networks (HTTP 500).
	// The controller drops the key by itself while Teleport is disabled.
	body := map[string]any{"enabled": t.Enabled}
	if t.SubnetCIDR != "" {
		body["subnet_cidr"] = t.SubnetCIDR
	}
	var out Teleport
	err := c.UpdateSetting(ctx, siteID, "teleport", id, body, &out)
	return out, err
}
