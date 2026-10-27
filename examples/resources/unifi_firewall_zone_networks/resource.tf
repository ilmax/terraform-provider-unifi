resource "unifi_firewall_zone_networks" "trusted_assignments" {
  site_id = var.unifi_site_id
  zone_id = unifi_firewall_zone.trusted.id

  network_ids = [
    unifi_network.lan.id,
    unifi_network.iot.id,
  ]
}
