---
page_title: "unifi_pending_devices Data Source"
---

# unifi_pending_devices (Data Source)

Lists all devices that have sent an inform to the controller but haven't been adopted into any site yet.

## Example Usage

```hcl
data "unifi_pending_devices" "all" {}

output "pending_macs" {
  value = [for d in data.unifi_pending_devices.all.devices : d.mac_address]
}
```

## Notes

- This is global (not scoped to a site) and GET-only in the real API — there's no way to adopt a device through this provider; adoption happens through the console UI or the `adoptDevice` API.
- Always returns `200` with a list, which may be empty (a console with no devices currently pending adoption).
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Read-Only

- `id` (String) Data source identifier.
- `devices` (List of Object) Devices currently pending adoption.

### `devices` Object

- `mac_address` (String) Device MAC address.
- `ip_address` (String) Device IP address.
- `model` (String) Device model.
- `state` (String) Adoption state.
- `firmware_version` (String) Firmware version, if reported.
- `firmware_updatable` (Boolean) Whether firmware can be updated before adoption.
- `supported` (Boolean) Whether this model is supported.
- `adoption_target_site_ids` (List of String) Site IDs this device can be adopted into.
- `features` (List of String) Device feature flags.
