---
page_title: "unifi_vpn_server Data Source"
---

# unifi_vpn_server (Data Source)

Reads a configured VPN server.

## Example Usage

```hcl
data "unifi_vpn_server" "example" {
  site_id       = var.unifi_site_id
  vpn_server_id = "vpn-server-123"
}
```

## Notes

- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- There is no lookup-by-name; `vpn_server_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `vpn_server_id` (String) VPN server identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) VPN server identifier.
- `name` (String) VPN server name.
- `type` (String) VPN server type.
- `enabled` (Boolean) Whether the VPN server is enabled.
- `origin` (String) Entity origin metadata.
