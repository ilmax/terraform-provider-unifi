// Package radius holds a hand-written response type for the RADIUS profile
// list endpoint (/v1/sites/{siteId}/radius/profiles), which is GET-only in
// the real API — see data_source_radius_profile.go.
package radius

type Metadata struct {
	Origin string `json:"origin"`
}

// Profile matches one entry of "Get all RADIUS profiles".
type Profile struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

// ListProfilesResponse matches the full paginated response envelope.
type ListProfilesResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Profile `json:"data"`
}
