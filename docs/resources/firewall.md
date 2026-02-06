---
page_title: "unifi_firewall Resource"
---

# unifi_firewall (Resource)

Manages an ACL firewall rule.

## Example Usage

```hcl
resource "unifi_firewall" "example" {
  site_id  = var.unifi_site_id
  type     = "IPV4"
  name     = "allow-dns"
  action   = "ALLOW"
  enabled  = true
  index    = 100

  source_filter_json = jsonencode({
    type = "SUBNETS"
    ipAddressesOrSubnets = ["10.0.0.0/24"]
  })

  destination_filter_json = jsonencode({
    type = "PORTS"
    ports = [53]
  })

  protocol_filter = ["UDP"]
}
```

## Schema

### Required

- `type` (String) ACL rule type. Example: `IPV4`, `MAC`.
- `name` (String) Rule name.
- `action` (String) Rule action. Example: `ALLOW`, `DENY`.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `description` (String) Rule description.
- `enabled` (Boolean) Enable or disable the rule.
- `index` (Number) Rule priority (lower is higher priority).
- `source_filter_json` (String) JSON-encoded source filter object.
- `destination_filter_json` (String) JSON-encoded destination filter object.
- `protocol_filter` (List of String) Protocol filter values.

### Read-Only

- `id` (String) ACL rule identifier.

## Import

```
<site_id>/<acl_rule_id>
```
