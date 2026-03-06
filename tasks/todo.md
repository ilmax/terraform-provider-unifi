# Todo

## Working Notes
- Adjust countries data source output:
- Replace `countries` nested list with a flat output list attribute.
- Keep pagination behavior unchanged.

## Plan
- [ ] Refactor `unifi_countries` schema/state output shape.
- [ ] Update tests for the new output attribute.
- [ ] Update docs/examples and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Data source fetches beyond the API default page size.
- Data source no longer exposes output under `countries`.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- In progress.
