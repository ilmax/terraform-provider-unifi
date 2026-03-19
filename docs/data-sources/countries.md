---
page_title: "unifi_countries Data Source"
---

# unifi_countries (Data Source)

Reads the full UniFi country catalog as a map keyed by country name.

The provider automatically paginates the `/v1/countries` endpoint to return
all available countries, not only the first page.

## Example Usage

```hcl
data "unifi_countries" "all" {}

locals {
  china_code = data.unifi_countries.all.countries["China"].code
}
```

## Schema

### Read-Only

- `id` (String) Constant data source identifier (`countries`).
- `countries` (Map of Object) Full country catalog keyed by country name.

Example access:

```hcl
data.unifi_countries.all.countries["China"].code
```

### `countries` Value Object

- `code` (String) Country code in ISO 3166-1 alpha-2 format.
