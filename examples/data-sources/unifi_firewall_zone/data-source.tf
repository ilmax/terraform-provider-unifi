data "unifi_firewall_zone" "example" {
  site_id = var.unifi_site_id
  name    = "trusted"
}
