data "unifi_switch_mc_lag_domain" "example" {
  site_id          = var.unifi_site_id
  mc_lag_domain_id = "mc-lag-123"
}
