---
page_title: "unifi_dns_policies Data Source"
---

# unifi_dns_policies (Data Source)

Lists DNS policies for a site, with optional filtering by type and domain.

## DNS Policy Types

The `type` filter and each returned policy `type` can currently be one of:

- `A_RECORD`
- `AAAA_RECORD`
- `CNAME_RECORD`
- `FORWARD_DOMAIN`
- `MX_RECORD`
- `SRV_RECORD`
- `TXT_RECORD`

## Example Usage

```hcl
data "unifi_dns_policies" "all" {
  site_id = var.unifi_site_id
}

data "unifi_dns_policies" "blocked_domain" {
  site_id = var.unifi_site_id
  type    = "FORWARD_DOMAIN"
  domain  = "lab.example"
}
```

## Schema

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `type` (String) Case-insensitive DNS policy type filter. See supported values above.
- `domain` (String) Case-insensitive DNS policy domain filter.

### Read-Only

- `id` (String) Data source identifier (site ID).
- `policies` (List of Object) Matching DNS policies.

### `policies` Object

- `id` (String) DNS policy identifier.
- `type` (String) DNS policy type. See supported values above.
- `enabled` (Boolean) DNS policy enabled state.
- `domain` (String) DNS policy domain (when present).
- `origin` (String) DNS policy origin metadata.
