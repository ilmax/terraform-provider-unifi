---
page_title: "unifi_vpn_site_to_site_tunnel Data Source"
---

# unifi_vpn_site_to_site_tunnel (Data Source)

Reads a configured VPN site-to-site tunnel.

## Example Usage

```hcl
data "unifi_vpn_site_to_site_tunnel" "example" {
  site_id   = var.unifi_site_id
  tunnel_id = "tunnel-123"
}
```

## Notes

- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- There is no lookup-by-name; `tunnel_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `tunnel_id` (String) VPN site-to-site tunnel identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) VPN site-to-site tunnel identifier.
- `name` (String) Tunnel name.
- `type` (String) Tunnel type.
- `origin` (String) Entity origin metadata.
