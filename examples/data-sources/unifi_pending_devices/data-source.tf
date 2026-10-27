data "unifi_pending_devices" "all" {}

output "pending_macs" {
  value = [for d in data.unifi_pending_devices.all.devices : d.mac_address]
}
