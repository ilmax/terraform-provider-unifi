data "unifi_vpn_site_to_site_tunnel" "example" {
  site_id   = var.unifi_site_id
  tunnel_id = "tunnel-123"
}
