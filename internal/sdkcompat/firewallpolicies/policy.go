// Package firewallpolicies holds hand-written response types for the
// firewall policy endpoints, matching the shapes verified against the
// official OpenAPI spec (developer.ui.com/network/v10.6.106/openapi.json).
//
// The request side is deliberately NOT modeled as typed structs here: a
// firewall policy's source/destination traffic filter is a deeply nested
// discriminated union (network/IP/subnet/range/port/MAC/domain/region/
// application/VPN-server/site-to-site-tunnel filters, each split again by
// IPv4/IPv6/both) — dozens of variants, far more than unifi_firewall_rule's
// four-variant filter (which resource_firewall_rule.go models natively as
// source_filter/destination_filter). resource_firewall_policy.go instead
// builds the request as map[string]any and exposes the filter as an opaque
// JSON string attribute — the escape hatch is still the right call here,
// where native modeling would mean dozens of near-duplicate optional fields.
package firewallpolicies

import "encoding/json"

// Policy matches the "Firewall policy" response schema, returned by
// create, update, and get — all three share this one shape.
type Policy struct {
	Id                    string                `json:"id"`
	Name                  string                `json:"name"`
	Description           string                `json:"description,omitempty"`
	Enabled               bool                  `json:"enabled"`
	LoggingEnabled        bool                  `json:"loggingEnabled"`
	IpProtocolScope       PolicyIPProtocolScope `json:"ipProtocolScope"`
	IpsecFilter           string                `json:"ipsecFilter,omitempty"`
	ConnectionStateFilter []string              `json:"connectionStateFilter,omitempty"`
	Index                 int64                 `json:"index"`
	Action                PolicyAction          `json:"action"`
	Source                PolicyEndpoint        `json:"source"`
	Destination           PolicyEndpoint        `json:"destination"`
	Metadata              *PolicyMetadata       `json:"metadata,omitempty"`
}

// PolicyIPProtocolScope matches "Firewall policy IP protocol scope" — a
// discriminated union on ipVersion (IPV4/IPV6/IPV4_AND_IPV6), each variant
// additionally carrying an optional protocolFilter (yet another deeply
// nested discriminated union of named protocols/presets/protocol numbers).
// Only ipVersion is modeled; protocolFilter is left unset, matching this
// resource's scope of not modeling protocol-level filtering within it.
type PolicyIPProtocolScope struct {
	IpVersion string `json:"ipVersion"`
}

// PolicyAction matches the "Firewall policy action" discriminated union.
// AllowReturnTraffic is only populated (by the API) when Type is "ALLOW".
type PolicyAction struct {
	Type               string `json:"type"`
	AllowReturnTraffic *bool  `json:"allowReturnTraffic,omitempty"`
}

// PolicyEndpoint matches "Firewall policy source"/"Firewall policy
// destination" — TrafficFilter is left as raw JSON since it's the deeply
// nested discriminated union described above.
type PolicyEndpoint struct {
	ZoneId        string          `json:"zoneId"`
	TrafficFilter json.RawMessage `json:"trafficFilter,omitempty"`
}

// PolicyMetadata mirrors the "origin" field pattern used by DNS policies
// and ACL rules (USER_DEFINED vs SYSTEM_DEFINED).
type PolicyMetadata struct {
	Origin string `json:"origin"`
}
