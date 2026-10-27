data "unifi_acl_rules" "all" {
  site_id = var.unifi_site_id
}

output "acl_rule_names" {
  value = [for r in data.unifi_acl_rules.all.acl_rules : r.name]
}
