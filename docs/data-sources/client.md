---
page_title: "unifi_client Data Source"
---

# unifi_client (Data Source)

Reads details about a connected client.

## Example Usage

```hcl
data "unifi_client" "example" {
  site_id   = var.unifi_site_id
  client_id = "client-123"
}
```

## Notes

- Clients can represent wired, wireless, VPN, or teleport connections depending on the UniFi Network API response.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `client_id` (String) Client identifier.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Client identifier.
- `name` (String) Client name.
- `type` (String) Client type.
- `connected_at` (String) RFC3339 timestamp when the client connected.
- `ip_address` (String) Client IP address.
- `access_type` (String) Access type reported by the UniFi API.
- `mac_address` (String) Client MAC address (wired/wireless clients).
- `uplink_device_id` (String) Uplink device identifier (wired/wireless clients).
