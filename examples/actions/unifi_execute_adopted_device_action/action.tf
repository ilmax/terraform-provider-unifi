action "unifi_execute_adopted_device_action" "restart_device" {
  config {
    site_id   = var.unifi_site_id
    device_id = "device-123"
    action    = "RESTART"
  }
}

resource "terraform_data" "trigger" {
  input = timestamp()

  lifecycle {
    action_trigger {
      events  = [after_update]
      actions = [action.unifi_execute_adopted_device_action.restart_device]
    }
  }
}
