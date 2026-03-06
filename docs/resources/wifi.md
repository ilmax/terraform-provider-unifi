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
  basic_data_rate_kbps_by_frequency_ghz = {
    rate_2_4_kbps = 6000
    rate_5_kbps   = 12000
  }

  client_filtering_policy = {
    action             = "ALLOW"
    mac_address_filter = ["aa:bb:cc:dd:ee:ff"]
  }

  blackout_schedule_configuration = {
    days = [
      {
        day  = "MONDAY"
        type = "ALL_DAY"
      }
    ]
  }
}
```

## Notes

- By default, access points broadcast SSIDs on all supported bands. Use `broadcasting_frequencies_ghz` to limit the broadcast to specific bands.
- Access points typically support four SSIDs per band, and up to eight when wireless meshing is disabled; too many SSIDs can reduce performance.
- MLO is a WiFi 7 feature that can use multiple bands for a single connection. UniFi documents support starting with UniFi Network 8.2.93 and WiFi 7 AP firmware 7.1.18, plus compatible clients.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

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
- `basic_data_rate_kbps_by_frequency_ghz` (Block) Basic data rate overrides by band.
- `client_filtering_policy` (Block) WiFi client filtering policy.
- `blackout_schedule_configuration` (Block) Blackout schedule per day.

### Read-Only

- `id` (String) WiFi broadcast identifier.

### `basic_data_rate_kbps_by_frequency_ghz` Block

- `rate_2_4_kbps` (Number) Basic data rate for 2.4 GHz in Kbps.
- `rate_5_kbps` (Number) Basic data rate for 5 GHz in Kbps.

### `client_filtering_policy` Block

- `action` (String) Filter policy action (for example `ALLOW` or `DENY`).
- `mac_address_filter` (List of String) MAC addresses used by the filter.

### `blackout_schedule_configuration` Block

- `days` (List of Object) Day schedule entries.

### `blackout_schedule_configuration.days` Object

- `day` (String) Day name (for example `MONDAY`).
- `type` (String) Day schedule type (for example `ALL_DAY`).

## Import

```sh
<site_id>/<wifi_broadcast_id>
```
