resource "unifi_dns_mx_record" "mail" {
  site_id            = var.unifi_site_id
  enabled            = true
  domain             = "example.com"
  mail_server_domain = "mail.example.com"
  priority           = 10
}
