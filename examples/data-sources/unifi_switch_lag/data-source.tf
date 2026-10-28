data "unifi_switch_lag" "example" {
  site_id = var.unifi_site_id
  lag_id  = "lag-123"
}
