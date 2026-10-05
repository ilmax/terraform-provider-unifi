// Package trafficmatching holds a hand-written response type for the
// traffic matching list endpoints, matching the shape verified against the
// official OpenAPI spec (developer.ui.com/network/v10.6.106/openapi.json).
//
// Like firewallpolicies, the request/response "items" array is left as raw
// JSON rather than modeled as typed structs: each item is itself a small
// discriminated union (IP_ADDRESS/IP_ADDRESS_RANGE/SUBNET for IPv4,
// IP_ADDRESS/SUBNET for IPv6, PORT_NUMBER/PORT_NUMBER_RANGE for ports).
// resource_traffic_matching_list.go exposes this as an opaque JSON array
// string attribute, the same escape-hatch pattern used throughout this
// provider for deeply nested discriminated unions.
package trafficmatching

import "encoding/json"

// List matches the "Traffic matching list" response schema, returned by
// create, update, and get — all three share this one shape.
type List struct {
	Id    string          `json:"id"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Items json.RawMessage `json:"items"`
}
