resource "unifi_traffic_matching_list" "office_subnets" {
  site_id = var.unifi_site_id
  name    = "office-subnets"
  type    = "IPV4_ADDRESSES"

  items_json = jsonencode([
    { type = "IP_ADDRESS", value = "192.168.1.5" },
    { type = "SUBNET", value = "10.0.0.0/24" },
    { type = "IP_ADDRESS_RANGE", start = "10.0.0.10", stop = "10.0.0.20" },
  ])
}

resource "unifi_traffic_matching_list" "web_ports" {
  site_id = var.unifi_site_id
  name    = "web-ports"
  type    = "PORTS"

  items_json = jsonencode([
    { type = "PORT_NUMBER", value = 443 },
    { type = "PORT_NUMBER_RANGE", start = 8000, stop = 8100 },
  ])
}
