// Package vpn holds hand-written response types for the VPN server and
// site-to-site tunnel list endpoints. Both are GET-only in the real API
// (no create/update/delete), so they're exposed as data sources, not
// resources — see data_source_vpn_server.go / data_source_vpn_site_to_site_tunnel.go.
package vpn

// Metadata matches the "User defined entity metadata" schema shared across
// many list responses in this API.
type Metadata struct {
	Origin string `json:"origin"`
}

// Server matches one entry of "Get all VPN servers" (/v1/sites/{siteId}/vpn/servers).
type Server struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Enabled  bool     `json:"enabled"`
	Metadata Metadata `json:"metadata"`
}

// ListServersResponse matches the full paginated response envelope.
type ListServersResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Server `json:"data"`
}

// SiteToSiteTunnel matches one entry of "Get all VPN site-to-site tunnels"
// (/v1/sites/{siteId}/vpn/site-to-site-tunnels).
type SiteToSiteTunnel struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Metadata Metadata `json:"metadata"`
}

// ListSiteToSiteTunnelsResponse matches the full paginated response envelope.
type ListSiteToSiteTunnelsResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []SiteToSiteTunnel `json:"data"`
}
