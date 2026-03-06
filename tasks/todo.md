# Todo

## Working Notes
- Align provider file naming with Terraform resource names for discoverability.
- Add `unifi_acl_rules` data source to list all ACL rules.

## Plan
- [x] Rename mismatched resource file(s), especially `unifi_firewall_rule`.
- [x] Implement `unifi_acl_rules` data source with pagination support.
- [x] Register the data source and add schema/unit tests.
- [x] Update docs/README/index and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_rule` implementation lives in a correspondingly named file.
- `unifi_acl_rules` lists all ACL rules for a site.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- Renamed `resource_firewall.go` to `resource_firewall_rule.go` to match `unifi_firewall_rule`.
- Added `unifi_acl_rules` data source with paginated retrieval of ACL rules.
- Registered the new data source and added schema/unit tests.
- Added docs for `unifi_acl_rules` and updated provider listings in `README.md` and `docs/index.md`.
- Verified with `go test ./...`.
