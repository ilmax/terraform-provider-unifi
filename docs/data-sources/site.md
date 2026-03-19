---
page_title: "unifi_site Data Source"
---

# unifi_site (Data Source)

Finds a UniFi site by its name and returns the corresponding `site_id`.

## Example Usage

```hcl
data "unifi_site" "home" {
  name = "Home"
}

resource "unifi_network" "lan" {
  site_id    = data.unifi_site.home.site_id
  name       = "LAN"
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 10
}
```

## Notes

- Site lookup is case-insensitive and ignores leading/trailing whitespace.
- If multiple sites share the same name, the data source returns an error so the configuration does not select an arbitrary site.

## Schema

### Required

- `name` (String) Site name to look up.

### Read-Only

- `id` (String) The resolved site ID.
- `site_id` (String) The resolved site ID.
- `host_id` (String) The host ID for the site.
- `description` (String) Site description.
- `timezone` (String) Site timezone.
