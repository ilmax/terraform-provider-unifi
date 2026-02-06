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
  hide_name     = false
  client_isolation_enabled              = false
  multicast_to_unicast_conversion_enabled = true
  broadcasting_frequencies_ghz          = ["2.4", "5"]
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
- `multicast_to_unicast_conversion_enabled` (Boolean) Convert multicast to unicast.
- `client_isolation_enabled` (Boolean) Enable client isolation.
- `hide_name` (Boolean) Hide SSID.
- `uapsd_enabled` (Boolean) Enable U-APSD.
- `broadcasting_frequencies_ghz` (List of String) Broadcast frequencies (STANDARD only).
- `mlo_enabled` (Boolean) Enable MLO (STANDARD only).
- `band_steering_enabled` (Boolean) Enable band steering (STANDARD only).
- `arp_proxy_enabled` (Boolean) Enable ARP proxy (STANDARD only).
- `bss_transition_enabled` (Boolean) Enable BSS transition (STANDARD only).

### Read-Only

- `id` (String) WiFi broadcast identifier.

## Import

```sh
<site_id>/<wifi_broadcast_id>
```
