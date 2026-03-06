# Todo

## Working Notes
- Add ordering support for ACL rules and firewall policies.

## Plan
- [x] Add `unifi_acl_rule_ordering` resource.
- [x] Add `unifi_firewall_policy_ordering` resource.
- [x] Register resources, update schema tests, and add docs.
- [x] Run go test ./... and commit per task.

## Acceptance Criteria
- ACL rule ordering can be managed via Terraform.
- Firewall policy ordering (by source/destination zone pair) can be managed via Terraform.
- Tests pass.

## Results
- Added `unifi_acl_rule_ordering` resource with import/read/write support.
- Added `unifi_firewall_policy_ordering` resource keyed by source/destination zone IDs.
- Registered resources in provider, added schema/unit tests, and documented usage/import behavior.
- Verified with `go test ./...`.
