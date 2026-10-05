---
page_title: "unifi_radius_profile Data Source"
---

# unifi_radius_profile (Data Source)

Reads a configured RADIUS profile.

## Example Usage

```hcl
data "unifi_radius_profile" "default" {
  site_id           = var.unifi_site_id
  radius_profile_id = "8d0da3df-9b9b-4376-92af-38e7f6fe2b88"
}
```

## Notes

- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- Every site has at least a `SYSTEM_DEFINED` "Default" RADIUS profile out of the box, with no manual configuration required.
- There is no lookup-by-name; `radius_profile_id` must be the exact ID.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `radius_profile_id` (String) RADIUS profile identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) RADIUS profile identifier.
- `name` (String) RADIUS profile name.
- `origin` (String) Entity origin metadata (e.g. `SYSTEM_DEFINED`, `DERIVED`, `USER_DEFINED`).
