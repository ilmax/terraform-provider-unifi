---
page_title: "unifi_dns_mx_record Resource"
---

# unifi_dns_mx_record (Resource)

Manages a UniFi DNS MX record policy.

## Example Usage

```hcl
resource "unifi_dns_mx_record" "mail" {
  site_id            = var.unifi_site_id
  enabled            = true
  domain             = "example.com"
  mail_server_domain = "mail.example.com"
  priority           = 10
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `mail_server_domain` (String) Mail server host for the record.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `domain` (String) Domain name for the record.
- `priority` (Number) MX priority; lower values are preferred.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
