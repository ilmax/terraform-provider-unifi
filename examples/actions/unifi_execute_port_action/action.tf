action "unifi_execute_port_action" "power_cycle_port" {
  config {
    site_id   = var.unifi_site_id
    device_id = "device-123"
    port_idx  = 1
    action    = "POWER_CYCLE"
  }
}

resource "terraform_data" "trigger" {
  input = timestamp()

  lifecycle {
    action_trigger {
      events  = [after_update]
      actions = [action.unifi_execute_port_action.power_cycle_port]
    }
  }
}
