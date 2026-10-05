---
page_title: "unifi_firewall_policy Resource"
---

# unifi_firewall_policy (Resource)

Manages a Zone Based Firewall policy: an allow/block/reject rule between two firewall zones.

## Example Usage

```hcl
resource "unifi_firewall_policy" "block_guest_to_internal" {
  site_id             = var.unifi_site_id
  name                = "block-guest-to-internal"
  action              = "BLOCK"
  source_zone_id      = unifi_firewall_zone.guest.zone_id
  destination_zone_id = unifi_firewall_zone.internal.zone_id
  logging_enabled     = true
}

resource "unifi_firewall_policy" "allow_internal_to_external" {
  site_id              = var.unifi_site_id
  name                 = "allow-internal-to-external"
  action               = "ALLOW"
  allow_return_traffic = true
  source_zone_id       = unifi_firewall_zone.internal.zone_id
  destination_zone_id  = unifi_firewall_zone.external.zone_id

  destination_traffic_filter_json = jsonencode({
    type  = "PORTS"
    ports = [443]
  })
}
```

## Notes

- Requires `unifi_firewall_zone` for `source_zone_id`/`destination_zone_id`, and (like real UniFi hardware) a gateway model that actually supports Zone Based Firewall — a UDM/UDR/UXG-class device. Classic gateways like the USG 3P never supported this feature.
- The traffic filter (matching specific networks/IPs/ports/etc. within a zone, rather than the whole zone) is a large discriminated union in the real API. `source_traffic_filter_json`/`destination_traffic_filter_json` take it as a raw JSON string — the same escape hatch `unifi_firewall_rule` uses for its filters — rather than modeling every filter variant as native attributes. Omit them to match all traffic to/from that zone.
- `allow_return_traffic` is only valid when `action = ALLOW`; setting it for `BLOCK`/`REJECT` is a plan-time error.
- Schedules are not yet supported: an unscheduled policy is always active, which is the common case.
- Priority among a site's policies is managed with `unifi_firewall_policy_ordering`, not here — `index` is read-only.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `name` (String) Policy name.
- `action` (String) `ALLOW`, `BLOCK`, or `REJECT`.
- `source_zone_id` (String) Source firewall zone ID.
- `destination_zone_id` (String) Destination firewall zone ID.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `description` (String) Policy description.
- `allow_return_traffic` (Boolean) Only valid when `action = ALLOW`: also creates a derived policy on the mirrored zone pair to allow return traffic. Omitting it sends `false`.
- `enabled` (Boolean) Enable or disable the policy. Defaults to `true`.
- `logging_enabled` (Boolean) Enable logging for matching traffic. Defaults to `false`.
- `ip_protocol_scope` (String) `IPV4`, `IPV6`, or `IPV4_AND_IPV6`. Defaults to `IPV4_AND_IPV6`.
- `ipsec_filter` (String) `MATCH_ENCRYPTED` or `MATCH_NOT_ENCRYPTED`. Omitted: matches all traffic regardless of IPsec encryption.
- `connection_state_filter` (Set of String) Any of `NEW`, `INVALID`, `ESTABLISHED`, `RELATED`. Omitted: matches all connection states. A set, not a list: the real API stores this unordered.
- `source_traffic_filter_json` (String) Raw JSON matching the API's source traffic filter shape. Omitted: matches all traffic from the source zone.
- `destination_traffic_filter_json` (String) Raw JSON matching the API's destination traffic filter shape. Omitted: matches all traffic to the destination zone.

### Read-Only

- `id` (String) Policy identifier.
- `index` (Number) Priority position among this site's firewall policies (lower = higher priority). Manage with `unifi_firewall_policy_ordering`, not here.
- `origin` (String) `USER_DEFINED`, `SYSTEM_DEFINED`, or `DERIVED` (auto-created by `allow_return_traffic` on another policy).

## Import

```sh
<site_id>/<firewall_policy_id>
```
