---
page_title: "unifi_firewall_policy_ordering Resource"
---

# unifi_firewall_policy_ordering (Resource)

Manages firewall policy ordering for a source/destination zone pair.

## Example Usage

```hcl
resource "unifi_firewall_policy_ordering" "internet_to_lan" {
  site_id                      = var.unifi_site_id
  source_firewall_zone_id      = data.unifi_firewall_zones.external.zones[0].id
  destination_firewall_zone_id = data.unifi_firewall_zones.internal.zones[0].id

  before_system_defined = [
    "11111111-1111-1111-1111-111111111111",
  ]

  after_system_defined = [
    "22222222-2222-2222-2222-222222222222",
  ]
}
```

## Notes

- UniFi tracks ordering separately for each source/destination firewall zone pair.
- This resource calls:
- `GET /v1/sites/{siteId}/firewall/policies/ordering`
- `PUT /v1/sites/{siteId}/firewall/policies/ordering`
- with `sourceFirewallZoneId` and `destinationFirewallZoneId` query parameters.
- Deleting this resource removes it from Terraform state only; it does not reset ordering in UniFi.

## Schema

### Required

- `source_firewall_zone_id` (String) Source firewall zone ID.
- `destination_firewall_zone_id` (String) Destination firewall zone ID.
- `before_system_defined` (List of String) Ordered firewall policy IDs placed before system-defined policies.
- `after_system_defined` (List of String) Ordered firewall policy IDs placed after system-defined policies.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Composite identifier in the format `<site_id>/<source_firewall_zone_id>/<destination_firewall_zone_id>`.

## Import

```sh
<site_id>/<source_firewall_zone_id>/<destination_firewall_zone_id>
```
