data "unifi_client" "example" {
  site_id   = var.unifi_site_id
  client_id = "client-123"
}

data "unifi_client" "by_mac" {
  site_id     = var.unifi_site_id
  mac_address = "aa:bb:cc:dd:ee:ff"
}
