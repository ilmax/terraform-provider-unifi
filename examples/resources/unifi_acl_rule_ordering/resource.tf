resource "unifi_acl_rule_ordering" "site_order" {
  site_id = var.unifi_site_id

  ordered_acl_rule_ids = [
    unifi_firewall_rule.allow_dns.id,
    unifi_firewall_rule.allow_ntp.id,
    unifi_firewall_rule.deny_all.id,
  ]
}
