---
page_title: "unifi_dns_cname_record Resource"
---

# unifi_dns_cname_record (Resource)

Manages a UniFi DNS CNAME record policy.

## Example Usage

```hcl
resource "unifi_dns_cname_record" "app" {
  site_id       = var.unifi_site_id
  enabled       = true
  domain        = "app.example.com"
  target_domain = "lb.internal.example.com"
  ttl_seconds   = 300
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `target_domain` (String) Canonical name target for the record.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `domain` (String) Domain name for the record.
- `ttl_seconds` (Number) DNS TTL in seconds.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
