---
page_title: "unifi_network Resource"
---

# unifi_network (Resource)

Manages a UniFi network.

## Example Usage

```hcl
resource "unifi_network" "example" {
  site_id    = var.unifi_site_id
  name       = "corp"
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 10

  ipv4_configuration_json = jsonencode({
    hostIpAddress = "10.0.0.1"
    prefixLength  = 24
    dhcpConfiguration = {
      mode = "SERVER"
      ipAddressRange = {
        start = "10.0.0.10"
        stop  = "10.0.0.200"
      }
      dnsServerIpAddressesOverride = ["1.1.1.1", "8.8.8.8"]
      leaseTimeSeconds             = 86400
      domainName                   = "corp.local"
    }
  })
}
```

## Schema

### Required

- `name` (String) Network name.
- `management` (String) Management type. Example: `GATEWAY`, `SWITCH`, `UNMANAGED`.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `enabled` (Boolean) Enable or disable the network.
- `vlan_id` (Number) VLAN ID.
- `ipv4_configuration_json` (String) JSON-encoded IPv4 configuration (includes DHCP and DNS overrides).
- `ipv6_configuration_json` (String) JSON-encoded IPv6 configuration.

### Read-Only

- `id` (String) Network identifier.

## Import

```
<site_id>/<network_id>
```
