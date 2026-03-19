# Todo

## Working Notes
- Add `unifi_site` data source to resolve a site ID from a site name.
- Sites are global, so this lookup should not depend on provider `site_id`.

## Plan
- [ ] Implement `unifi_site` data source with paginated site lookup by name.
- [ ] Register it and add schema/unit tests.
- [ ] Update docs/README/index and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_site` resolves `site_id` by site name.
- Duplicate site names return a clear error.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- In progress.
