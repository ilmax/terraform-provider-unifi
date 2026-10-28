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
