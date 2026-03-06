# Terraform Provider for UniFi

Terraform provider for UniFi Cloud (Site Manager API) using `github.com/ilmax/unifi-client-go`.

## Authentication

Only API key authentication is supported.

```hcl
terraform {
  required_providers {
    unifi = {
      source  = "ilmax/unifi"
      version = "~> 0.3"
    }
  }
}
```

```hcl
provider "unifi" {
  api_key = var.unifi_api_key
  site_id = var.unifi_site_id
  # api_url = "https://api.ui.com" # optional override
  # allow_insecure = true          # optional, for local/test controllers
}
```

## Local Controller Example

To connect to a local UniFi OS console (for example a UDR7), point `api_url` at
the local integration endpoint and (if needed) enable `allow_insecure` for
self-signed certificates:

```hcl
provider "unifi" {
  api_key        = var.unifi_api_key
  site_id        = var.unifi_site_id
  api_url        = "https://<console-address>/proxy/network/integration"
  allow_insecure = true
}
```

## Debugging

Enable debug/trace logging with Terraform environment variables. Trace logging
includes pretty-printed request/response bodies (use only for debugging; it can
include sensitive data).

```sh
export TF_LOG=DEBUG
export TF_LOG_PROVIDER=TRACE
export TF_LOG_PATH=./terraform.log
```

## Resources

- `unifi_network`
- `unifi_wifi` (changes force replacement)
- `unifi_firewall_rule`
- `unifi_firewall_zone`

## Data Sources

- `unifi_client` (lookup by `client_id` or `mac_address`)
- `unifi_device`
- `unifi_firewall_zone`
- `unifi_firewall_zones`
- `unifi_wan`

## Actions

- `unifi_execute_port_action`
- `unifi_execute_adopted_device_action`

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
