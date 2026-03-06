# Todo

## Working Notes
- Add DNS policies support as focused Terraform entities:
- `unifi_dns_policy` resource
- `unifi_dns_policy` data source (single lookup)
- `unifi_dns_policies` data source (list/filter)

## Plan
- [ ] Implement `unifi_dns_policy` resource (CRUD + import + state mapping).
- [ ] Implement `unifi_dns_policy` and `unifi_dns_policies` data sources.
- [ ] Register provider entries and extend schema/unit tests.
- [ ] Add docs for resource/data sources and update provider index/README.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- DNS policies can be managed via Terraform with create/read/update/delete/import.
- DNS policy can be looked up as a single data source and listed via collection data source.
- Provider/docs expose new DNS entities.
- Tests pass.

## Results
- In progress.
