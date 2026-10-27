resource "unifi_dns_srv_record" "ldap" {
  site_id       = var.unifi_site_id
  enabled       = true
  domain        = "example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv.example.com"
  port          = 389
  priority      = 10
  weight        = 20
}
