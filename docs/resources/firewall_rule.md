---
page_title: "unifi_firewall_rule Resource"
---

# unifi_firewall_rule (Resource)

Manages an ACL firewall rule.

## Example Usage

```hcl
resource "unifi_firewall_rule" "allow_dns" {
  site_id = var.unifi_site_id
  type    = "IPV4"
  name    = "allow-dns"
  action  = "ALLOW"
  enabled = true

  source_filter = {
    type                    = "IP_ADDRESSES_OR_SUBNETS"
    ip_addresses_or_subnets = ["10.0.0.0/24"]
  }

  destination_filter = {
    type        = "PORTS"
    port_filter = [53]
  }

  protocol_filter = ["UDP"]
}

resource "unifi_firewall_rule" "block_guest_device" {
  site_id           = var.unifi_site_id
  type              = "MAC"
  name              = "block-guest-device"
  action            = "BLOCK"
  enabled           = true
  network_id_filter = unifi_network.guest.id

  source_filter = {
    type          = "MAC_ADDRESSES"
    mac_addresses = ["aa:bb:cc:dd:ee:ff"]
  }
}
```

## Notes

- `source_filter`/`destination_filter` model the real API's discriminated union of match criteria as native attributes rather than a raw JSON string: one shared shape covers every variant (`ip_addresses_or_subnets`, `network_ids`, `port_filter`, `mac_addresses`, `prefix_length`), and only the field(s) matching the chosen `type` may be set — the provider validates the rest are omitted.
- For an `IPV4` rule, `type` is `IP_ADDRESSES_OR_SUBNETS`, `NETWORKS`, or `PORTS`. `port_filter` is required when `type = PORTS`, and optional as an extra restriction on the other two — but only when `protocol_filter` is also set on the rule (the real API rejects a port filter without one).
- For a `MAC` rule, `type` is always `MAC_ADDRESSES`; `prefix_length` (1-48) is optional and narrows the match to a MAC address prefix instead of full addresses.
- `network_id_filter` is required by the API when `type = MAC`; unused for `IPV4`.
- Rules are evaluated by index (lower numbers are evaluated first). For explicit ordering management, use `unifi_acl_rule_ordering`.
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
- `source_filter` (Object) Traffic source filter. Omit to match all traffic. See `source_filter`/`destination_filter` below.
- `destination_filter` (Object) Traffic destination filter. Omit to match all traffic. See `source_filter`/`destination_filter` below.
- `protocol_filter` (Set of String) Protocols this rule applies to, e.g. `TCP`, `UDP`. When omitted, applies to all protocols. A set, not a list: the real API stores this unordered.
- `network_id_filter` (String) Network ID this ACL rule applies to. Required by the API when type is MAC; unused for IPV4.

### Read-Only

- `id` (String) ACL rule identifier.

### `source_filter`/`destination_filter` Object

- `type` (String, Required) `IP_ADDRESSES_OR_SUBNETS`, `NETWORKS`, or `PORTS` for an IPV4 rule; `MAC_ADDRESSES` for a MAC rule.
- `ip_addresses_or_subnets` (List of String) IP addresses or CIDR subnets to match. Required when `type = IP_ADDRESSES_OR_SUBNETS`; must be omitted otherwise.
- `network_ids` (List of String) Network IDs to match. Required when `type = NETWORKS`; must be omitted otherwise.
- `port_filter` (List of Number) Ports (1-65535) to match. Required when `type = PORTS`; an optional extra restriction when `type = IP_ADDRESSES_OR_SUBNETS` or `NETWORKS` (needs `protocol_filter` set on the rule too); must be omitted when `type = MAC_ADDRESSES`.
- `mac_addresses` (List of String) MAC addresses to match. Required when `type = MAC_ADDRESSES`; must be omitted otherwise.
- `prefix_length` (Number) MAC address prefix length (1-48). Only valid when `type = MAC_ADDRESSES`; omit for a full-address match.

## Import

```sh
<site_id>/<acl_rule_id>
```
