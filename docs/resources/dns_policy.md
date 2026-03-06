---
page_title: "unifi_dns_policy Resource"
---

# unifi_dns_policy (Resource)

Manages a DNS policy.

## Example Usage

```hcl
resource "unifi_dns_policy" "block_ads" {
  site_id = var.unifi_site_id
  type    = "BLOCK"
  enabled = true
}
```

## Notes

- This resource uses `POST/GET/PUT/DELETE /v1/sites/{siteId}/dns-policies`.
- `domain` is exposed as read-only state from UniFi API responses.

## Schema

### Required

- `type` (String) DNS policy type.
- `enabled` (Boolean) Enable or disable the DNS policy.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) DNS policy identifier.
- `domain` (String) DNS policy domain (when present).
- `origin` (String) DNS policy origin metadata.

## Import

```sh
<site_id>/<dns_policy_id>
```
