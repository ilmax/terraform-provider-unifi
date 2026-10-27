data "unifi_vpn_server" "example" {
  site_id       = var.unifi_site_id
  vpn_server_id = "vpn-server-123"
}
