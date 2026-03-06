---
page_title: "Provider: unifi"
---

# unifi Provider

The UniFi provider manages UniFi Cloud (Site Manager API) resources via API key authentication.

## Authentication

Only API key authentication is supported.

```hcl
provider "unifi" {
  api_key = var.unifi_api_key
  site_id = var.unifi_site_id
  # api_url = "https://api.ui.com" # optional override
  # allow_insecure = true          # optional, for local/test controllers
}
```

## Configuration Reference

- `api_key` (String, Required, Sensitive) API key for UniFi Cloud.
- `site_id` (String, Optional) Default site identifier used by resources.
- `api_url` (String, Optional) Override the UniFi API base URL.
- `user_agent` (String, Optional) Custom user agent string.
- `allow_insecure` (Boolean, Optional) Skip TLS certificate verification.

## Resources

- `unifi_acl_rule_ordering`
- `unifi_dns_policy`
- `unifi_firewall_policy_ordering`
- `unifi_firewall_rule`
- `unifi_firewall_zone`
- `unifi_network`
- `unifi_wifi`

## Data Sources

- `unifi_dns_policy`
- `unifi_dns_policies`
- `unifi_client`
- `unifi_device`
- `unifi_firewall_zone`
- `unifi_firewall_zones`
- `unifi_wan`

## Actions

- `unifi_execute_port_action`
- `unifi_execute_adopted_device_action`
