---
page_title: "unifi_device Resource"
---

# unifi_device (Resource)

Read-only resource that reads details about an adopted device.

## Example Usage

```hcl
resource "unifi_device" "example" {
  site_id   = var.unifi_site_id
  device_id = "device-123"
}
```

## Schema

### Required

- `device_id` (String) Device identifier.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Device identifier.
- `name` (String) Device name.
- `mac_address` (String) MAC address.
- `ip_address` (String) IP address.
- `model` (String) Device model.
- `state` (String) Device state.
- `firmware_version` (String) Firmware version.
- `firmware_updatable` (Boolean) Firmware update available.
- `supported` (Boolean) Device support status.

## Import

```
<site_id>/<device_id>
```
