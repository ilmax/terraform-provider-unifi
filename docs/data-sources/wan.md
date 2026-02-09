---
page_title: "unifi_wan Data Source"
---

# unifi_wan (Data Source)

Reads WAN interface details.

## Example Usage

```hcl
data "unifi_wan" "example" {
  site_id = var.unifi_site_id
  wan_id  = "wan-123"
}
```

## Notes

- WAN interfaces are commonly associated with the External zone to represent internet-facing traffic in zone-based policies.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `wan_id` (String) WAN interface identifier.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) WAN identifier.
- `name` (String) WAN interface name.
