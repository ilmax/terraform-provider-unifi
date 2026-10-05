// Package vouchers holds hand-written request/response types for the
// hotspot voucher endpoints, matching the shapes verified against the
// official OpenAPI spec (developer.ui.com/network/v10.6.106/openapi.json).
// There is no dedicated Go SDK coverage for this endpoint family, so these
// mirror the same house style as internal/sdkcompat/acl_rules and friends.
package vouchers

import "time"

// CreateVoucherRequest is the body for POST /v1/sites/{siteId}/hotspot/vouchers.
// Count is intentionally omitted here — the resource always creates exactly
// one voucher per Terraform resource instance (see resource_hotspot_voucher.go),
// letting Terraform's own count/for_each handle generating many, rather than
// exposing the API's own batch-of-up-to-1000 knob directly.
type CreateVoucherRequest struct {
	Name                 string `json:"name"`
	Count                int64  `json:"count"`
	TimeLimitMinutes     int64  `json:"timeLimitMinutes"`
	AuthorizedGuestLimit int64  `json:"authorizedGuestLimit,omitempty"`
	DataUsageLimitMBytes int64  `json:"dataUsageLimitMBytes,omitempty"`
	RxRateLimitKbps      int64  `json:"rxRateLimitKbps,omitempty"`
	TxRateLimitKbps      int64  `json:"txRateLimitKbps,omitempty"`
}

// CreateVoucherResponse wraps the created voucher(s) — always exactly one
// element here since CreateVoucherRequest.Count is always 1.
type CreateVoucherResponse struct {
	Vouchers []VoucherDetails `json:"vouchers"`
}

// VoucherDetails matches the "Hotspot voucher details" schema, returned by
// both create and the single-voucher GET.
type VoucherDetails struct {
	Id                   string     `json:"id"`
	Name                 string     `json:"name"`
	Code                 string     `json:"code"`
	CreatedAt            time.Time  `json:"createdAt"`
	ActivatedAt          *time.Time `json:"activatedAt,omitempty"`
	ExpiresAt            *time.Time `json:"expiresAt,omitempty"`
	Expired              bool       `json:"expired"`
	AuthorizedGuestCount int64      `json:"authorizedGuestCount"`
	AuthorizedGuestLimit int64      `json:"authorizedGuestLimit,omitempty"`
	TimeLimitMinutes     int64      `json:"timeLimitMinutes"`
	DataUsageLimitMBytes int64      `json:"dataUsageLimitMBytes,omitempty"`
	RxRateLimitKbps      int64      `json:"rxRateLimitKbps,omitempty"`
	TxRateLimitKbps      int64      `json:"txRateLimitKbps,omitempty"`
}

// ListVouchersResponse matches the "Hotspot voucher detail page" schema.
type ListVouchersResponse struct {
	Offset     int64            `json:"offset"`
	Limit      int64            `json:"limit"`
	Count      int64            `json:"count"`
	TotalCount int64            `json:"totalCount"`
	Data       []VoucherDetails `json:"data"`
}
