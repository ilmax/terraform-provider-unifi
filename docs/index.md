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
- `site_id` (String, Optional) Default site identifier used by site-scoped resources, data sources, and actions. A local `site_id` overrides this provider value.
- `api_url` (String, Optional) Override the UniFi API base URL.
- `user_agent` (String, Optional) Custom user agent string.
- `allow_insecure` (Boolean, Optional) Skip TLS certificate verification.

## Site ID Precedence

For every site-scoped resource, data source, and action:

1. a local `site_id` wins if set on that block
2. otherwise the provider-level `site_id` is used
3. if neither is set, the provider returns an error

## Resources

- `unifi_acl_rule_ordering`
- `unifi_dns_a_record`
- `unifi_dns_aaaa_record`
- `unifi_dns_cname_record`
- `unifi_dns_forward_domain_policy`
- `unifi_dns_mx_record`
- `unifi_dns_srv_record`
- `unifi_dns_txt_record`
- `unifi_firewall_policy`
- `unifi_firewall_policy_ordering`
- `unifi_firewall_rule`
- `unifi_firewall_zone`
- `unifi_firewall_zone_networks`
- `unifi_hotspot_voucher`
- `unifi_network`
- `unifi_traffic_matching_list`
- `unifi_wifi`

## Data Sources

- `unifi_countries`
- `unifi_acl_rules`
- `unifi_device_tag`
- `unifi_dns_policy`
- `unifi_dns_policies`
- `unifi_dpi_application`
- `unifi_dpi_category`
- `unifi_client`
- `unifi_device`
- `unifi_firewall_zone`
- `unifi_firewall_zones`
- `unifi_pending_devices`
- `unifi_radius_profile`
- `unifi_site`
- `unifi_switch_lag`
- `unifi_switch_mc_lag_domain`
- `unifi_switch_stack`
- `unifi_vpn_server`
- `unifi_vpn_site_to_site_tunnel`
- `unifi_wan`

## Actions

- `unifi_execute_port_action`
- `unifi_execute_adopted_device_action`
