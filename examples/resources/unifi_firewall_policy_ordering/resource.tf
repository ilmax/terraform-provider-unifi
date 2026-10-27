resource "unifi_firewall_policy_ordering" "internet_to_lan" {
  site_id                      = var.unifi_site_id
  source_firewall_zone_id      = data.unifi_firewall_zones.external.zones[0].id
  destination_firewall_zone_id = data.unifi_firewall_zones.internal.zones[0].id

  before_system_defined = [
    "11111111-1111-1111-1111-111111111111",
  ]

  after_system_defined = [
    "22222222-2222-2222-2222-222222222222",
  ]
}
