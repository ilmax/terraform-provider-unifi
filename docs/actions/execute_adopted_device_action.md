---
page_title: "unifi_execute_adopted_device_action Action"
---

# unifi_execute_adopted_device_action (Action)

Executes an action on an adopted UniFi device. Actions are experimental in
Terraform and require Terraform v1.14 or later.

## Example Usage

```hcl
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
```

## Arguments

- `device_id` (String, Required) Adopted device identifier.
- `action` (String, Required) Device action to execute. Supported: `RESTART`.
- `site_id` (String, Optional) Site identifier (defaults to provider `site_id`).
- `timeout_seconds` (Number, Optional) Timeout in seconds for the action request (default: 1800).
