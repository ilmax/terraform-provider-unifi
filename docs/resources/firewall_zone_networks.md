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
- Only a GATEWAY-managed network can belong to a zone at all — the real API rejects any other management type with a misleading `"Configured network does not exist"` error, even though the network genuinely exists.
- Deleting this resource unassigns every network it listed via the real API (the real API falls back to the site's Internal zone, since a GATEWAY network's zone assignment is mandatory once Zone Based Firewall is enabled), rather than only forgetting them in Terraform state — a zone whose `network_ids` still names a network Terraform is about to destroy next can't itself be deleted afterward.

## Schema

### Required

- `zone_id` (String) Firewall zone identifier.
- `network_ids` (List of String) Network IDs to attach to the zone.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Resource identifier in the form `<site_id>/<zone_id>`.

## Import

```sh
<site_id>/<zone_id>
```
