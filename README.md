# Terraform Provider for UniFi

Terraform provider for UniFi Cloud (Site Manager API) using `github.com/ilmax/unifi-client-go`.

## Authentication

Only API key authentication is supported.

```hcl
provider "unifi" {
  api_key = var.unifi_api_key
  site_id = var.unifi_site_id
}
```

## Resources

- `unifi_device` (read-only)
- `unifi_network`
- `unifi_wifi` (changes force replacement)
- `unifi_firewall`
- `unifi_wan` (read-only)

## Import

Resources that require a `site_id` use this import format:

```
<site_id>/<resource_id>
```

Example:

```
terraform import unifi_network.example "site-123/network-456"
```
