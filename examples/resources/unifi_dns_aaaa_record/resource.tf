resource "unifi_dns_aaaa_record" "app" {
  site_id      = var.unifi_site_id
  enabled      = true
  domain       = "app.example.com"
  ipv6_address = "2001:db8::10"
  ttl_seconds  = 300
}
