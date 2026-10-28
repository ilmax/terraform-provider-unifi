resource "unifi_dns_forward_domain_policy" "lab" {
  site_id    = var.unifi_site_id
  enabled    = true
  domain     = "lab.example.com"
  ip_address = "192.168.1.53"
}
