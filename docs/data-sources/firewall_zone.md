---
page_title: "unifi_firewall_zone Data Source"
---

# unifi_firewall_zone (Data Source)

Reads a single firewall zone by name.

## Example Usage

```hcl
data "unifi_firewall_zone" "example" {
  site_id = var.unifi_site_id
  name    = "trusted"
}
```

## Schema

### Required

- `name` (String) Firewall zone name.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Firewall zone identifier.
- `zone_id` (String) Firewall zone identifier.
- `name` (String) Firewall zone name.
- `network_ids` (List of String) Attached network IDs.
- `origin` (String) Zone origin metadata.
