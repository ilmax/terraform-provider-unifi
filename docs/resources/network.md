---
page_title: "unifi_network Resource"
---

# unifi_network (Resource)

Manages a UniFi network.

## Example Usage

```hcl
resource "unifi_network" "example" {
  site_id    = var.unifi_site_id
  name       = "corp"
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 10
}
```

## Schema

### Required

- `name` (String) Network name.
- `management` (String) Management type. Example: `GATEWAY`, `SWITCH`, `UNMANAGED`.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `enabled` (Boolean) Enable or disable the network.
- `vlan_id` (Number) VLAN ID.

### Read-Only

- `id` (String) Network identifier.

## Import

```
<site_id>/<network_id>
```
