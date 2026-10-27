resource "unifi_wifi" "example" {
  site_id       = var.unifi_site_id
  name          = "corp-wifi"
  type          = "STANDARD"
  enabled       = true
  security_type = "WPA2_PERSONAL"
  network_type  = "NATIVE"
  hide_name     = false
  client_isolation_enabled              = false
  multicast_to_unicast_conversion_enabled = true
  broadcasting_frequencies_ghz          = ["2.4", "5"]
  basic_data_rate_kbps_by_frequency_ghz = {
    rate_2_4_kbps = 6000
    rate_5_kbps   = 12000
  }

  client_filtering_policy = {
    action             = "ALLOW"
    mac_address_filter = ["aa:bb:cc:dd:ee:ff"]
  }

  blackout_schedule_configuration = {
    days = [
      {
        day  = "MONDAY"
        type = "ALL_DAY"
      }
    ]
  }
}
