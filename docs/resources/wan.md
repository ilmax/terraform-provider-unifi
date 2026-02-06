---
page_title: "unifi_wan Resource"
---

# unifi_wan (Resource)

Read-only resource for WAN interfaces.

## Example Usage

```hcl
resource "unifi_wan" "example" {
  site_id = var.unifi_site_id
  wan_id  = "wan-123"
}
```

## Schema

### Required

- `wan_id` (String) WAN interface identifier.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) WAN identifier.
- `name` (String) WAN interface name.

## Import

```sh
<site_id>/<wan_id>
```
