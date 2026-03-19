---
page_title: "unifi_dns_a_record Resource"
---

# unifi_dns_a_record (Resource)

Manages a UniFi DNS A record policy.

## Example Usage

```hcl
resource "unifi_dns_a_record" "app" {
  site_id      = var.unifi_site_id
  enabled      = true
  domain       = "app.example.com"
  ipv4_address = "192.168.1.10"
  ttl_seconds  = 300
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `ipv4_address` (String) IPv4 address returned for the record.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `domain` (String) Domain name for the record.
- `ttl_seconds` (Number) DNS TTL in seconds.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
