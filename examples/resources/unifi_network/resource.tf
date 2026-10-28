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
