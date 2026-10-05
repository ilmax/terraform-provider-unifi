// Package switching holds hand-written response types for the switch LAG,
// MC-LAG domain, and switch stack endpoints
// (/v1/sites/{siteId}/switching/{lags,mc-lag-domains,switch-stacks}), all
// GET-only in the real API — see data_source_switch_lag.go,
// data_source_switch_mc_lag_domain.go, and data_source_switch_stack.go.
// Unlike the traffic matching lists / firewall policies in sibling
// packages, these shapes have no discriminated unions, so they're modeled
// natively rather than through the opaque-JSON escape hatch.
package switching

type Metadata struct {
	Origin string `json:"origin"`
}

// Lag matches one entry of "Get all switch LAGs"
// (/v1/sites/{siteId}/switching/lags) and the single-item GET response —
// deliberately sparse: it identifies a LAG's existence and type, not its
// member ports (those live nested under McLagDomain/SwitchStack instead).
type Lag struct {
	Id       string   `json:"id"`
	Type     string   `json:"type"`
	Metadata Metadata `json:"metadata"`
}

type ListLagsResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Lag `json:"data"`
}

// LagMember matches one entry of McLagLocal.members.
type LagMember struct {
	DeviceId string  `json:"deviceId"`
	PortIdxs []int64 `json:"portIdxs"`
}

// McLagLocal matches one entry of McLagDomain.lags.
type McLagLocal struct {
	Id       string      `json:"id"`
	Members  []LagMember `json:"members"`
	Metadata Metadata    `json:"metadata"`
}

// McLagPeer matches one entry of McLagDomain.peers.
type McLagPeer struct {
	DeviceId     string  `json:"deviceId"`
	LinkPortIdxs []int64 `json:"linkPortIdxs"`
	Role         string  `json:"role"`
}

// McLagDomain matches the single-item GET response for
// /v1/sites/{siteId}/switching/mc-lag-domains/{mcLagDomainId}.
type McLagDomain struct {
	Id       string       `json:"id"`
	Name     string       `json:"name"`
	Lags     []McLagLocal `json:"lags"`
	Peers    []McLagPeer  `json:"peers"`
	Metadata Metadata     `json:"metadata"`
}

// SwitchStackLagMember matches one entry of SwitchStackLagLocal.members.
type SwitchStackLagMember struct {
	PortIdxs       []int64 `json:"portIdxs"`
	UnitId         int64   `json:"unitId"`
	UnitMacAddress string  `json:"unitMacAddress"`
}

// SwitchStackLagLocal matches one entry of SwitchStack.lags.
type SwitchStackLagLocal struct {
	Id       string                 `json:"id"`
	Members  []SwitchStackLagMember `json:"members"`
	Metadata Metadata               `json:"metadata"`
}

// SwitchStackUnit matches one entry of SwitchStack.units.
type SwitchStackUnit struct {
	Id         int64  `json:"id"`
	MacAddress string `json:"macAddress"`
	Order      *int64 `json:"order,omitempty"`
	Role       string `json:"role,omitempty"`
}

// SwitchStack matches the single-item GET response for
// /v1/sites/{siteId}/switching/switch-stacks/{switchStackId}.
type SwitchStack struct {
	Id       string                `json:"id"`
	Name     string                `json:"name"`
	DeviceId string                `json:"deviceId,omitempty"`
	Lags     []SwitchStackLagLocal `json:"lags"`
	Units    []SwitchStackUnit     `json:"units"`
	Metadata Metadata              `json:"metadata"`
}
