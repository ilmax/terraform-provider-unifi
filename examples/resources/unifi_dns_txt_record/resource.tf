resource "unifi_dns_txt_record" "spf" {
  site_id = var.unifi_site_id
  enabled = true
  domain  = "example.com"
  text    = "v=spf1 include:example.com ~all"
}
