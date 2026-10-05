---
page_title: "unifi_dpi_category Data Source"
---

# unifi_dpi_category (Data Source)

Reads a Deep Packet Inspection category identifier, e.g. for referencing in traffic rules.

## Example Usage

```hcl
data "unifi_dpi_category" "instant_messengers" {
  category_id = 0
}
```

## Notes

- This is global reference data (not scoped to a site) and GET-only in the real API, so it's a data source rather than a resource.
- Like `unifi_dpi_application`, `category_id` is a small integer, not a UUID.
- There is no lookup-by-name; `category_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `category_id` (Number) DPI category identifier.

### Read-Only

- `id` (String) Data source identifier (string form of `category_id`).
- `name` (String) Category name.
