---
page_title: "unifi_wifi Resource"
---

# unifi_wifi (Resource)

Manages a WiFi broadcast (SSID). Changes to this resource force replacement.

## Example Usage

```hcl
resource "unifi_wifi" "example" {
  site_id       = var.unifi_site_id
  name          = "corp-wifi"
  type          = "STANDARD"
  enabled       = true
  security_type = "WPA2_PERSONAL"
  network_type  = "NATIVE"
}
```

## Schema

### Required

- `name` (String) WiFi name.
- `type` (String) Broadcast type. Example: `STANDARD`, `IOT_OPTIMIZED`.
- `security_type` (String) Security configuration type. Example: `OPEN`, `WPA2_PERSONAL`.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).
- `enabled` (Boolean) Enable or disable the broadcast.
- `network_type` (String) Network reference type, such as `NATIVE` or `SPECIFIC`.

### Read-Only

- `id` (String) WiFi broadcast identifier.

## Import

```
<site_id>/<wifi_broadcast_id>
```
