# Todo

## Working Notes
- Align provider file naming with Terraform resource names for discoverability.
- Add `unifi_acl_rules` data source to list all ACL rules.

## Plan
- [ ] Rename mismatched resource file(s), especially `unifi_firewall_rule`.
- [ ] Implement `unifi_acl_rules` data source with pagination support.
- [ ] Register the data source and add schema/unit tests.
- [ ] Update docs/README/index and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_rule` implementation lives in a correspondingly named file.
- `unifi_acl_rules` lists all ACL rules for a site.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- In progress.
