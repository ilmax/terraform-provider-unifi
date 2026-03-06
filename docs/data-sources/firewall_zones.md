---
page_title: "unifi_firewall_zones Data Source"
---

# unifi_firewall_zones (Data Source)

Lists firewall zones for a site, with an optional name filter.

## Example Usage

```hcl
data "unifi_firewall_zones" "all" {
  site_id = var.unifi_site_id
}

data "unifi_firewall_zones" "external" {
  site_id = var.unifi_site_id
  name    = "External"
}
```

## Schema

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `name` (String) Case-insensitive firewall zone name filter.

### Read-Only

- `id` (String) Data source identifier (site ID).
- `zones` (List of Object) Matching firewall zones.

### `zones` Object

- `id` (String) Firewall zone identifier.
- `name` (String) Firewall zone name.
- `network_ids` (List of String) Attached network IDs.
- `origin` (String) Zone origin metadata.
