---
page_title: "unifi_dns_srv_record Resource"
---

# unifi_dns_srv_record (Resource)

Manages a UniFi DNS SRV record policy.

## Example Usage

```hcl
resource "unifi_dns_srv_record" "ldap" {
  site_id       = var.unifi_site_id
  enabled       = true
  domain        = "example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv.example.com"
  port          = 389
  priority      = 10
  weight        = 20
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `service` (String) SRV service label such as `_ldap`.
- `protocol` (String) SRV protocol label such as `_tcp`.
- `server_domain` (String) Hostname serving the service.
- `port` (Number) Service port.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `domain` (String) Domain name for the record.
- `priority` (Number) SRV priority; lower values are preferred.
- `weight` (Number) SRV weight among records with equal priority.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
