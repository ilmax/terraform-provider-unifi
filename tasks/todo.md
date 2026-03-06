# Todo

## Working Notes
- Add DNS policies support as focused Terraform entities:
- `unifi_dns_policy` resource
- `unifi_dns_policy` data source (single lookup)
- `unifi_dns_policies` data source (list/filter)

## Plan
- [x] Implement `unifi_dns_policy` resource (CRUD + import + state mapping).
- [x] Implement `unifi_dns_policy` and `unifi_dns_policies` data sources.
- [x] Register provider entries and extend schema/unit tests.
- [x] Add docs for resource/data sources and update provider index/README.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- DNS policies can be managed via Terraform with create/read/update/delete/import.
- DNS policy can be looked up as a single data source and listed via collection data source.
- Provider/docs expose new DNS entities.
- Tests pass.

## Results
- Added `unifi_dns_policy` resource with CRUD/import support and state mapping for `domain` and `origin`.
- Added `unifi_dns_policy` and `unifi_dns_policies` data sources with optional type/domain filters for list reads.
- Registered DNS entities in provider resource/data source registries and extended schema/unit tests.
- Added DNS docs in `README.md`, provider index, and dedicated resource/data source pages.
- Verified with `go test ./...`.
