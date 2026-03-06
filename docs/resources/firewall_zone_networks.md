---
page_title: "unifi_firewall_zone_networks Resource"
---

# unifi_firewall_zone_networks (Resource)

Manages network assignments for an existing firewall zone.

## Example Usage

```hcl
resource "unifi_firewall_zone_networks" "trusted_assignments" {
  site_id = var.unifi_site_id
  zone_id = unifi_firewall_zone.trusted.id

  network_ids = [
    unifi_network.lan.id,
    unifi_network.iot.id,
  ]
}
```

## Notes

- This resource updates only the zone's `network_ids`.
- The zone name/lifecycle should be managed by `unifi_firewall_zone`.
- Deleting this resource removes it from Terraform state only; it does not change zone assignments in UniFi.

## Schema

### Required

- `zone_id` (String) Firewall zone identifier.
- `network_ids` (List of String) Network IDs to attach to the zone.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Resource identifier in the form `<site_id>/<zone_id>`.

## Import

```sh
<site_id>/<zone_id>
```
