# Todo

## Working Notes
- Add `unifi_site` data source to resolve a site ID from a site name.
- Sites are global, so this lookup should not depend on provider `site_id`.

## Plan
- [x] Implement `unifi_site` data source with paginated site lookup by name.
- [x] Register it and add schema/unit tests.
- [x] Update docs/README/index and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_site` resolves `site_id` by site name.
- Duplicate site names return a clear error.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- Added `unifi_site` data source to resolve `site_id` from a site name.
- Implemented paginated site listing using `pageSize` and `nextToken`.
- Added clear errors for missing and duplicate site name matches.
- Registered the data source, added schema/unit tests, and documented usage in `README.md` and `docs/index.md`.
- Verified with `go test ./...`.
