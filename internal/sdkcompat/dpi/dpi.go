// Package dpi holds hand-written response types for the (non-site-scoped,
// global) DPI application and category reference-data endpoints
// (/v1/dpi/applications, /v1/dpi/categories), which are GET-only in the
// real API — see data_source_dpi_application.go / data_source_dpi_category.go.
package dpi

// Application matches one entry of "Get all DPI applications". Unlike most
// other list endpoints in this API, its id is a small integer, not a UUID.
type Application struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// ListApplicationsResponse matches the full paginated response envelope.
type ListApplicationsResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Application `json:"data"`
}

// Category matches one entry of "Get all DPI categories". Like Application,
// its id is a small integer, not a UUID.
type Category struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// ListCategoriesResponse matches the full paginated response envelope.
type ListCategoriesResponse struct {
	Offset, Limit, Count, TotalCount int64
	Data                             []Category `json:"data"`
}
