resource "unifi_dns_a_record" "app" {
  site_id      = var.unifi_site_id
  enabled      = true
  domain       = "app.example.com"
  ipv4_address = "192.168.1.10"
  ttl_seconds  = 300
}
