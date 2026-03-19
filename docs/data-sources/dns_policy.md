---
page_title: "unifi_dns_policy Data Source"
---

# unifi_dns_policy (Data Source)

Reads a single DNS policy by policy ID.

## Example Usage

```hcl
data "unifi_dns_policy" "policy" {
  site_id   = var.unifi_site_id
  policy_id = var.dns_policy_id
}
```

## Schema

## DNS Policy Types

The returned `type` attribute can currently be one of:

- `A_RECORD`
- `AAAA_RECORD`
- `CNAME_RECORD`
- `FORWARD_DOMAIN`
- `MX_RECORD`
- `SRV_RECORD`
- `TXT_RECORD`

### Required

- `policy_id` (String) DNS policy identifier.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) DNS policy identifier.
- `type` (String) DNS policy type. See supported values above.
- `enabled` (Boolean) DNS policy enabled state.
- `domain` (String) DNS policy domain (when present).
- `origin` (String) DNS policy origin metadata.
