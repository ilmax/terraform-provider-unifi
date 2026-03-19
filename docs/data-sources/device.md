---
page_title: "unifi_device Data Source"
---

# unifi_device (Data Source)

Reads details about an adopted device.

## Example Usage

```hcl
data "unifi_device" "example" {
  site_id   = var.unifi_site_id
  device_id = "device-123"
}
```

## Notes

- Device details include model-specific capabilities, such as switch ports on gateways/switches or radio details on access points, when available.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `device_id` (String) Device identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

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
