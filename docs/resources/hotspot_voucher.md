---
page_title: "unifi_hotspot_voucher Resource"
---

# unifi_hotspot_voucher (Resource)

Manages a UniFi Hotspot guest WiFi voucher.

## Example Usage

```hcl
resource "unifi_hotspot_voucher" "guest_day_pass" {
  site_id                 = var.unifi_site_id
  name                    = "front-desk-day-pass"
  time_limit_minutes      = 1440
  authorized_guest_limit  = 1
  data_usage_limit_mbytes = 2048
  rx_rate_limit_kbps      = 5000
  tx_rate_limit_kbps      = 5000
}
```

## Notes

- The API has no update endpoint for vouchers — they're immutable once created, matching how a printed/issued voucher code works in practice. Every attribute forces replacement.
- The real API supports creating up to 1000 vouchers in a single call (`count`), but this resource always creates exactly one, so that Terraform's own `count`/`for_each` handles multiplicity the way it does for every other resource.
- `authorized_guest_limit`, `data_usage_limit_mbytes`, `rx_rate_limit_kbps`, and `tx_rate_limit_kbps` stay unset (unlimited) when omitted.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `name` (String) Voucher note. May be duplicated across multiple vouchers — it's a label, not an identifier.
- `time_limit_minutes` (Number) How long, in minutes, the voucher grants access starting from when the first guest authorizes with it.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `authorized_guest_limit` (Number) Limit on how many different guests can authorize with this same voucher.
- `data_usage_limit_mbytes` (Number) Data usage limit in megabytes.
- `rx_rate_limit_kbps` (Number) Download rate limit in kilobits per second.
- `tx_rate_limit_kbps` (Number) Upload rate limit in kilobits per second.

### Read-Only

- `id` (String) Voucher identifier.
- `code` (String) Secret code guests use to activate the voucher on the Hotspot portal.
- `authorized_guest_count` (Number) How many guests have used this voucher to authorize network access so far.
- `created_at` (String) When the voucher was created (RFC 3339).
- `activated_at` (String) When the first guest authorized with this voucher (RFC 3339). Null until then.
- `expires_at` (String) When the voucher stops granting access (RFC 3339). Null until activated.
- `expired` (Boolean) Whether the voucher has expired.

## Import

```sh
<site_id>/<voucher_id>
```
