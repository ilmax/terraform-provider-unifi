// Package devicetags holds a hand-written response type for the device tag
// list endpoint (/v1/sites/{siteId}/device-tags), which is GET-only in the
// real API — see data_source_device_tag.go.
package devicetags

type Metadata struct {
	Origin string `json:"origin"`
}

// Tag matches one entry of "Get all device tags".
type Tag struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	DeviceIds []string `json:"deviceIds"`
	Metadata  Metadata `json:"metadata"`
}

// ListTagsResponse matches the full paginated response envelope.
type ListTagsResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Tag `json:"data"`
}
