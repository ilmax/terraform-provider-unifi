---
page_title: "unifi_dns_policies Data Source"
---

# unifi_dns_policies (Data Source)

Lists DNS policies for a site, with optional filtering by type and domain.

## Example Usage

```hcl
data "unifi_dns_policies" "all" {
  site_id = var.unifi_site_id
}

data "unifi_dns_policies" "blocked_domain" {
  site_id = var.unifi_site_id
  type    = "BLOCK"
  domain  = "ads.example"
}
```

## Schema

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `type` (String) Case-insensitive DNS policy type filter.
- `domain` (String) Case-insensitive DNS policy domain filter.

### Read-Only

- `id` (String) Data source identifier (site ID).
- `policies` (List of Object) Matching DNS policies.

### `policies` Object

- `id` (String) DNS policy identifier.
- `type` (String) DNS policy type.
- `enabled` (Boolean) DNS policy enabled state.
- `domain` (String) DNS policy domain (when present).
- `origin` (String) DNS policy origin metadata.
