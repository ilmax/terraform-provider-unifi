# Terraform Provider for UniFi

Terraform provider for UniFi Cloud (Site Manager API) using `github.com/ilmax/unifi-client-go`.

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

## Resources

- `unifi_device` (read-only)
- `unifi_network`
- `unifi_wifi` (changes force replacement)
- `unifi_firewall`
- `unifi_firewall_zone`
- `unifi_wan` (read-only)

## Versioning

Provider releases use SemVer for the provider version and include the UniFi Network API
version in build metadata. Tag format:

```text
vX.Y.Z+unifi.A.B.C
```

Example:

```text
v0.3.0+unifi.8.2.93
```

`X.Y.Z` is the provider version. `+unifi.A.B.C` records the UniFi Network API version
the release targets; build metadata does not affect SemVer precedence.

## Import

Resources that require a `site_id` use this import format:

```sh
<site_id>/<resource_id>
```

Example:

```sh
terraform import unifi_network.example "site-123/network-456"
```
