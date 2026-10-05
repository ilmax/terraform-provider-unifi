---
page_title: "unifi_switch_lag Data Source"
---

# unifi_switch_lag (Data Source)

Reads a switch port LAG (link aggregation group).

## Example Usage

```hcl
data "unifi_switch_lag" "example" {
  site_id = var.unifi_site_id
  lag_id  = "lag-123"
}
```

## Notes

- Requires an adopted switch to have anything to read.
- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- Deliberately sparse: it identifies a LAG's existence and type, not its member ports — those are only available nested under `unifi_switch_mc_lag_domain`/`unifi_switch_stack`.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `lag_id` (String) Switch LAG identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Switch LAG identifier.
- `type` (String) LAG type.
- `origin` (String) Entity origin metadata.
