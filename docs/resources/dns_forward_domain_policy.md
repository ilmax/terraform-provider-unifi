---
page_title: "unifi_dns_forward_domain_policy Resource"
---

# unifi_dns_forward_domain_policy (Resource)

Manages a UniFi DNS forward-domain policy.

## Example Usage

```hcl
resource "unifi_dns_forward_domain_policy" "lab" {
  site_id    = var.unifi_site_id
  enabled    = true
  domain     = "lab.example.com"
  ip_address = "192.168.1.53"
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `ip_address` (String) DNS server IP address used for forwarding.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `domain` (String) Domain suffix to forward.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
