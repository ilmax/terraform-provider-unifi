---
page_title: "unifi_acl_rules Data Source"
---

# unifi_acl_rules (Data Source)

Lists ACL rules for a site.

## Example Usage

```hcl
data "unifi_acl_rules" "all" {
  site_id = var.unifi_site_id
}

output "acl_rule_names" {
  value = [for r in data.unifi_acl_rules.all.acl_rules : r.name]
}
```

## Schema

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Data source identifier (site ID).
- `acl_rules` (List of Object) ACL rules for the site.

### `acl_rules` Object

- `id` (String) ACL rule ID.
- `type` (String) ACL rule type.
- `name` (String) ACL rule name.
- `description` (String) ACL rule description.
- `action` (String) ACL rule action.
- `enabled` (Boolean) ACL rule enabled state.
- `index` (Number) ACL rule index (lower is higher priority).
- `origin` (String) ACL rule origin metadata.
