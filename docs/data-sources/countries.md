---
page_title: "unifi_countries Data Source"
---

# unifi_countries (Data Source)

Reads the full UniFi country catalog (ISO code + display name).

The provider automatically paginates the `/v1/countries` endpoint to return
all available countries, not only the first page.

## Example Usage

```hcl
data "unifi_countries" "all" {}

locals {
  country_code_by_name = {
    for c in data.unifi_countries.all.items :
    c.name => c.code
  }
}
```

## Schema

### Read-Only

- `id` (String) Constant data source identifier (`countries`).
- `items` (List of Object) Full list of countries.

### `items` Object

- `code` (String) Country code in ISO 3166-1 alpha-2 format.
- `name` (String) Country display name.
