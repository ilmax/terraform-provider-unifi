---
page_title: "unifi_traffic_matching_list Resource"
---

# unifi_traffic_matching_list (Resource)

Manages a reusable named list of IPv4 addresses, IPv6 addresses, or ports.

## Example Usage

```hcl
resource "unifi_traffic_matching_list" "office_subnets" {
  site_id = var.unifi_site_id
  name    = "office-subnets"
  type    = "IPV4_ADDRESSES"

  items_json = jsonencode([
    { type = "IP_ADDRESS", value = "192.168.1.5" },
    { type = "SUBNET", value = "10.0.0.0/24" },
    { type = "IP_ADDRESS_RANGE", start = "10.0.0.10", stop = "10.0.0.20" },
  ])
}

resource "unifi_traffic_matching_list" "web_ports" {
  site_id = var.unifi_site_id
  name    = "web-ports"
  type    = "PORTS"

  items_json = jsonencode([
    { type = "PORT_NUMBER", value = 443 },
    { type = "PORT_NUMBER_RANGE", start = 8000, stop = 8100 },
  ])
}
```

## Notes

- Referenced from traffic filters (e.g. `unifi_firewall_policy`'s `source_traffic_filter_json`/`destination_traffic_filter_json`) instead of repeating the same match criteria inline.
- Each entry in `items_json` is itself a small discriminated union, so it's exposed as a raw JSON array string — the same opaque-JSON escape hatch used elsewhere in this provider — rather than modeled as native attributes:
  - `IPV4_ADDRESSES`: `IP_ADDRESS` (`value`), `IP_ADDRESS_RANGE` (`start`/`stop`), `SUBNET` (`value`, CIDR).
  - `IPV6_ADDRESSES`: `IP_ADDRESS` (`value`), `SUBNET` (`value`, CIDR). No range type.
  - `PORTS`: `PORT_NUMBER` (`value`), `PORT_NUMBER_RANGE` (`start`/`stop`).
- Unlike `unifi_firewall_policy`, this endpoint is standalone (not nested under `/firewall/`), so it isn't gated behind Zone Based Firewall or any adopted gateway model.
- `name` and `items_json` are mutable in place via the real API's `PUT`. `type` requires replacement, since changing it would mean incompatible item semantics.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `name` (String) List name.
- `type` (String) `IPV4_ADDRESSES`, `IPV6_ADDRESSES`, or `PORTS`. Changing this requires replacement, since it changes the shape every item in `items_json` must match.
- `items_json` (String) Raw JSON array of items matching the API's shape for the chosen type.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) List identifier.

## Import

```sh
<site_id>/<traffic_matching_list_id>
```
