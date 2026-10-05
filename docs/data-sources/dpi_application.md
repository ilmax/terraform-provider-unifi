---
page_title: "unifi_dpi_application Data Source"
---

# unifi_dpi_application (Data Source)

Reads a Deep Packet Inspection application identifier, e.g. for referencing in traffic rules.

## Example Usage

```hcl
data "unifi_dpi_application" "icq" {
  application_id = 3
}
```

## Notes

- This is global reference data (not scoped to a site) and GET-only in the real API, so it's a data source rather than a resource.
- Unlike most other identifiers in this API, `application_id` is a small integer, not a UUID.
- There is no lookup-by-name; `application_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `application_id` (Number) DPI application identifier.

### Read-Only

- `id` (String) Data source identifier (string form of `application_id`).
- `name` (String) Application name.
