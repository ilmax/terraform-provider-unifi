---
page_title: "unifi_countries Data Source"
---

# unifi_countries (Data Source)

Reads the full UniFi country catalog (ISO code + display name).

## Example Usage

```hcl
data "unifi_countries" "all" {}

locals {
  country_code_by_name = {
    for c in data.unifi_countries.all.countries :
    c.name => c.code
  }
}
```

## Schema

### Read-Only

- `id` (String) Constant data source identifier (`countries`).
- `countries` (List of Object) Full list of countries.

### `countries` Object

- `code` (String) Country code in ISO 3166-1 alpha-2 format.
- `name` (String) Country display name.
