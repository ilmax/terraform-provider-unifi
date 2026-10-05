// Package pendingdevices holds a hand-written response type for the
// (non-site-scoped, global) pending-adoption device list endpoint
// (/v1/pending-devices), which is GET-only in the real API — see
// data_source_pending_devices.go.
package pendingdevices

// Device matches one entry of "Get all pending devices" — a device that has
// sent an inform to the controller but hasn't been adopted into any site
// yet (see test-infra/unifi/README.md's "Adopted device fleet" section for
// a concrete example of a device passing through this state).
type Device struct {
	MacAddress            string   `json:"macAddress"`
	IpAddress             string   `json:"ipAddress"`
	Model                 string   `json:"model"`
	State                 string   `json:"state"`
	FirmwareVersion       string   `json:"firmwareVersion,omitempty"`
	FirmwareUpdatable     bool     `json:"firmwareUpdatable"`
	Supported             bool     `json:"supported"`
	AdoptionTargetSiteIds []string `json:"adoptionTargetSiteIds"`
	Features              []string `json:"features"`
}

// ListPendingDevicesResponse matches the full paginated response envelope.
type ListPendingDevicesResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Device `json:"data"`
}
