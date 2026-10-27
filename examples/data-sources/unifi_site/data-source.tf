data "unifi_site" "home" {
  name = "Home"
}

resource "unifi_network" "lan" {
  site_id    = data.unifi_site.home.site_id
  name       = "LAN"
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 10
}
