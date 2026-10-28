data "unifi_dns_policy" "policy" {
  site_id   = var.unifi_site_id
  policy_id = var.dns_policy_id
}
