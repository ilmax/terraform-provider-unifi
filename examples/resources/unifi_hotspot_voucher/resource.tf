resource "unifi_hotspot_voucher" "guest_day_pass" {
  site_id                 = var.unifi_site_id
  name                    = "front-desk-day-pass"
  time_limit_minutes      = 1440
  authorized_guest_limit  = 1
  data_usage_limit_mbytes = 2048
  rx_rate_limit_kbps      = 5000
  tx_rate_limit_kbps      = 5000
}
