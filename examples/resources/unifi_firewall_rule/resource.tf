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
