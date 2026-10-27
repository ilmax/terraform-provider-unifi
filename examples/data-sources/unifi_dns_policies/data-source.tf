data "unifi_dns_policies" "all" {
  site_id = var.unifi_site_id
}

data "unifi_dns_policies" "blocked_domain" {
  site_id = var.unifi_site_id
  type    = "FORWARD_DOMAIN"
  domain  = "lab.example"
}
