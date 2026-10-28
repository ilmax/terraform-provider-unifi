resource "unifi_dns_cname_record" "app" {
  site_id       = var.unifi_site_id
  enabled       = true
  domain        = "app.example.com"
  target_domain = "lb.internal.example.com"
  ttl_seconds   = 300
}
