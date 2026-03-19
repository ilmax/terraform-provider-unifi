---
page_title: "unifi_firewall_zone Data Source"
---

# unifi_firewall_zone (Data Source)

Reads a single firewall zone by zone ID.

## Example Usage

```hcl
data "unifi_firewall_zone" "example" {
  site_id = var.unifi_site_id
  zone_id = "ed426e98-252f-4ff2-982b-d475007f085f"
}
```

## Schema

### Required

- `zone_id` (String) Firewall zone identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Firewall zone identifier.
- `name` (String) Firewall zone name.
- `network_ids` (List of String) Attached network IDs.
- `origin` (String) Zone origin metadata.
