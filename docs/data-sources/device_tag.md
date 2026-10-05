---
page_title: "unifi_device_tag Data Source"
---

# unifi_device_tag (Data Source)

Reads a device tag grouping adopted devices under a shared label.

## Example Usage

```hcl
data "unifi_device_tag" "example" {
  site_id       = var.unifi_site_id
  device_tag_id = "tag-123"
}
```

## Notes

- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- There is no lookup-by-name; `device_tag_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `device_tag_id` (String) Device tag identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Device tag identifier.
- `name` (String) Tag name.
- `device_ids` (List of String) IDs of devices carrying this tag.
- `origin` (String) Entity origin metadata.
