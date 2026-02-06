---
page_title: "unifi_firewall_zone Resource"
---

# unifi_firewall_zone (Resource)

Manages a custom firewall zone.

## Example Usage

```hcl
resource "unifi_firewall_zone" "example" {
  site_id     = var.unifi_site_id
  name        = "trusted"
  network_ids = [unifi_network.example.id]
}
```

## Schema

### Required

- `name` (String) Firewall zone name.
- `network_ids` (List of String) Network IDs attached to this zone.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Firewall zone identifier.

## Import

```sh
<site_id>/<firewall_zone_id>
```
