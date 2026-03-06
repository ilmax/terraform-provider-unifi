---
page_title: "unifi_acl_rule_ordering Resource"
---

# unifi_acl_rule_ordering (Resource)

Manages ACL rule ordering for a site.

## Example Usage

```hcl
resource "unifi_acl_rule_ordering" "site_order" {
  site_id = var.unifi_site_id

  ordered_acl_rule_ids = [
    unifi_firewall_rule.allow_dns.id,
    unifi_firewall_rule.allow_ntp.id,
    unifi_firewall_rule.deny_all.id,
  ]
}
```

## Notes

- This resource controls ordering via `PUT /v1/sites/{siteId}/acl-rules/ordering`.
- Deleting this resource removes it from Terraform state only; it does not reset ACL ordering in UniFi.

## Schema

### Required

- `ordered_acl_rule_ids` (List of String) Ordered ACL rule IDs.

### Optional

- `site_id` (String) Site identifier (defaults to provider `site_id`).

### Read-Only

- `id` (String) Resource identifier (site ID).

## Import

```sh
<site_id>
```
