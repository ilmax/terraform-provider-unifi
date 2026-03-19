---
page_title: "unifi_execute_port_action Action"
---

# unifi_execute_port_action (Action)

Executes an action on a UniFi switch port. Actions are experimental in Terraform
and require Terraform v1.14 or later.

## Example Usage

```hcl
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
```

## Arguments

- `device_id` (String, Required) Device identifier that owns the port.
- `port_idx` (Number, Required) Port index on the device.
- `action` (String, Required) Port action to execute. Supported: `POWER_CYCLE`.
- `site_id` (String, Optional) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.
- `timeout_seconds` (Number, Optional) Timeout in seconds for the action request (default: 1800).
