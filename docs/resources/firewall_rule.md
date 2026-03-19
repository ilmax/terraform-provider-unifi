---
page_title: "unifi_firewall_rule Resource"
---

# unifi_firewall_rule (Resource)

Manages an ACL firewall rule.

## Example Usage

```hcl
resource "unifi_firewall_rule" "example" {
  site_id  = var.unifi_site_id
  type     = "IPV4"
  name     = "allow-dns"
  action   = "ALLOW"
  enabled  = true
  index    = 100

  source_filter_json = jsonencode({
    type                 = "SUBNETS"
    ipAddressesOrSubnets = ["10.0.0.0/24"]
  })

  destination_filter_json = jsonencode({
    type  = "PORTS"
    ports = [53]
  })

  protocol_filter = ["UDP"]
}
```

## Notes

- Rules are evaluated by index (lower numbers are evaluated first). Use `source_filter_json`, `destination_filter_json`, and `protocol_filter` together to scope the traffic.
- For explicit ordering management, use `unifi_acl_rule_ordering`.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `type` (String) ACL rule type. Example: `IPV4`, `MAC`.
- `name` (String) Rule name.
- `action` (String) Rule action. Example: `ALLOW`, `DENY`.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `description` (String) Rule description.
- `enabled` (Boolean) Enable or disable the rule.
- `index` (Number) Rule priority (lower is higher priority).
- `source_filter_json` (String) JSON-encoded source filter object.
- `destination_filter_json` (String) JSON-encoded destination filter object.
- `protocol_filter` (List of String) Protocol filter values.

### Read-Only

- `id` (String) ACL rule identifier.

## Import

```sh
<site_id>/<acl_rule_id>
```
