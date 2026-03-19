---
page_title: "unifi_dns_txt_record Resource"
---

# unifi_dns_txt_record (Resource)

Manages a UniFi DNS TXT record policy.

## Example Usage

```hcl
resource "unifi_dns_txt_record" "spf" {
  site_id = var.unifi_site_id
  enabled = true
  domain  = "example.com"
  text    = "v=spf1 include:example.com ~all"
}
```

## Schema

### Required

- `enabled` (Boolean) Whether the DNS policy is active.
- `text` (String) TXT record value.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `domain` (String) Domain name for the record.

### Read-Only

- `id` (String) DNS policy identifier.
- `origin` (String) UniFi metadata origin.

## Import

```sh
<site_id>/<dns_policy_id>
```
