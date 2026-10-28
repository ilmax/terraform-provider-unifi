data "unifi_firewall_zones" "all" {
  site_id = var.unifi_site_id
}

data "unifi_firewall_zones" "external" {
  site_id = var.unifi_site_id
  name    = "External"
}
