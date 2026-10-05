---
page_title: "unifi_switch_stack Data Source"
---

# unifi_switch_stack (Data Source)

Reads a stack of physically linked switches acting as one logical switch.

## Example Usage

```hcl
data "unifi_switch_stack" "example" {
  site_id         = var.unifi_site_id
  switch_stack_id = "stack-123"
}
```

## Notes

- Requires stackable switch hardware to have anything to read.
- This endpoint is GET-only in the real API (no create/update/delete), so it's a data source rather than a resource.
- The UniFi Network API provides configuration and real-time status for sites, devices, and clients. It authenticates using an `X-API-Key` header (set `api_key` in the provider).

## Schema

### Required

- `switch_stack_id` (String) Switch stack identifier.

### Optional

- `site_id` (String) Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.

### Read-Only

- `id` (String) Switch stack identifier.
- `name` (String) Switch stack name.
- `device_id` (String) Primary device ID for this stack, if reported.
- `origin` (String) Entity origin metadata.
- `lags` (List of Object) LAGs local to this stack.
- `units` (List of Object) Physical switch units making up this stack.

### `lags` Object

- `id` (String) LAG identifier.
- `origin` (String) Entity origin metadata.
- `members` (List of Object) LAG member ports.

### `lags.members` Object

- `port_idxs` (List of Number) Port indexes making up this member.
- `unit_id` (Number) Stack unit ID this member port belongs to.
- `unit_mac_address` (String) MAC address of the stack unit this member port belongs to.

### `units` Object

- `id` (Number) Stack unit ID.
- `mac_address` (String) Unit MAC address.
- `order` (Number) Unit's position in the stack, if reported.
- `role` (String) Unit role (e.g. master/member), if reported.
