---
page_title: "unifi_firewall_zone Resource"
---

# unifi_firewall_zone (Resource)

Manages a custom firewall zone.

## Example Usage

```hcl
resource "unifi_firewall_zone" "example" {
  site_id = var.unifi_site_id
  name    = "trusted"
}
```

## Notes

- Firewall zones group interfaces (VLANs, WANs, VPNs) so you can manage policies between zones instead of individual networks.
- UniFi ships with built-in zones (such as External, Internal, Gateway, VPN, Hotspot, DMZ). Custom zones are supported, and interfaces can belong to one zone at a time.
- Zone-based policies are directional, and can also be applied within a zone when needed.
- Use `unifi_firewall_zone_networks` to manage network assignments in a zone.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `name` (String) Firewall zone name.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Firewall zone identifier.

## Import

```sh
<site_id>/<firewall_zone_id>
```
