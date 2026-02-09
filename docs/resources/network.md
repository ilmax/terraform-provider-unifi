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

  ipv4_configuration {
    host_ip_address = "10.0.0.1"
    prefix_length   = 24

    dhcp_configuration {
      mode = "SERVER"
      ip_address_range {
        start = "10.0.0.10"
        stop  = "10.0.0.200"
      }
      dns_servers      = ["1.1.1.1", "8.8.8.8"]
      lease_time_seconds               = 86400
      domain_name                      = "corp.local"
    }
  }
}
```

## Notes

- UniFi firewall zones group interfaces (VLANs, WANs, VPNs) to manage policy between zones. Networks can belong to a single zone at a time, and are placed into a built-in zone by default if you do not specify `zone_id`.
- Zone-based policies are directional; traffic is evaluated separately in each direction and can also be filtered within the same zone.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `name` (String) Network name.
- `management` (String) Management type. Example: `GATEWAY`, `SWITCH`, `UNMANAGED`.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `enabled` (Boolean) Enable or disable the network.
- `vlan_id` (Number) VLAN ID.
- `zone_id` (String) Firewall zone ID (gateway networks).
- `device_id` (String) Switch device ID (switch networks).
- `isolation_enabled` (Boolean) Whether the network is isolated.
- `cellular_backup_enabled` (Boolean) Allow cellular backup.
- `internet_access_enabled` (Boolean) Allow internet access.
- `multicast_dns_enable` (Boolean) Enable multicast DNS forwarding.
- `dhcp_guarding` (Block) DHCP guarding configuration.
- `ipv4_configuration` (Block) IPv4 configuration, including DHCP/DNS settings.
- `ipv6_configuration` (Block) IPv6 configuration.

### Read-Only

- `id` (String) Network identifier.

### `dhcp_guarding` Block

- `trusted_dhcp_server_ip_addresses` (List of String) Trusted DHCP server IPs.

### `ipv4_configuration` Block

- `auto_scale_enabled` (Boolean) Enable auto-scaling subnet sizing.
- `host_ip_address` (String) Gateway IP.
- `prefix_length` (Number) CIDR prefix length.
- `additional_host_ip_subnets` (List of String) Additional subnets.
- `dhcp_configuration` (Block) DHCP settings.

### `ipv4_configuration.dhcp_configuration` Block

- `mode` (String) DHCP mode (for example `SERVER` or `RELAY`).
- `ip_address_range` (Block) DHCP pool range.
- `gateway_ip_address_override` (String) Gateway override.
- `dns_servers` (List of String) DNS servers.
- `lease_time_seconds` (Number) Lease time in seconds.
- `domain_name` (String) DNS search domain.

### `ipv4_configuration.dhcp_configuration.ip_address_range` Block

- `start` (String) Range start.
- `stop` (String) Range end.

### `ipv6_configuration` Block

- `interface_type` (String) Interface type.
- `prefix_delegation_wan_interface_id` (String) WAN interface for PD.
- `dns_servers` (List of String) DNS servers.
- `additional_host_ip_subnets` (List of String) Additional subnets.
- `client_address_assignment` (Block) Client addressing.
- `router_advertisement` (Block) Router advertisement.

### `ipv6_configuration.client_address_assignment` Block

- `slaac_enabled` (Boolean) Enable SLAAC.
- `dhcp_configuration` (Block) DHCPv6 settings.

### `ipv6_configuration.client_address_assignment.dhcp_configuration` Block

- `ip_address_suffix_range` (Block) DHCPv6 suffix range.
- `lease_time_seconds` (Number) Lease time in seconds.

### `ipv6_configuration.client_address_assignment.dhcp_configuration.ip_address_suffix_range` Block

- `start` (String) Range start.
- `stop` (String) Range end.

### `ipv6_configuration.router_advertisement` Block

- `priority` (String) Router advertisement priority.

## Import

```sh
<site_id>/<network_id>
```
