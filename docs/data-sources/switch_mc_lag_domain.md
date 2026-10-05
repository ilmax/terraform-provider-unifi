---
page_title: "unifi_switch_mc_lag_domain Data Source"
---

# unifi_switch_mc_lag_domain (Data Source)

Reads an MC-LAG (multi-chassis link aggregation) domain spanning two switches.

## Example Usage

```hcl
data "unifi_switch_mc_lag_domain" "example" {
  site_id          = var.unifi_site_id
  mc_lag_domain_id = "mc-lag-123"
}
```

## Notes

- Requires MC-LAG capable switch hardware to have anything to read.
- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `mc_lag_domain_id` (String) MC-LAG domain identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) MC-LAG domain identifier.
- `name` (String) MC-LAG domain name.
- `origin` (String) Entity origin metadata.
- `lags` (List of Object) LAGs local to this domain.
- `peers` (List of Object) Peer switches forming this MC-LAG domain.

### `lags` Object

- `id` (String) LAG identifier.
- `origin` (String) Entity origin metadata.
- `members` (List of Object) LAG member ports.

### `lags.members` Object

- `device_id` (String) Device ID this member port belongs to.
- `port_idxs` (List of Number) Port indexes making up this member.

### `peers` Object

- `device_id` (String) Peer device ID.
- `link_port_idxs` (List of Number) Port indexes used for the inter-chassis link.
- `role` (String) Peer role in the domain.
