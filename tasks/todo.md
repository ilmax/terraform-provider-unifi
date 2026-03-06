# Todo

## Working Notes
- Adjust countries data source output:
- Replace `countries` nested list with a flat output list attribute.
- Keep pagination behavior unchanged.

## Plan
- [x] Refactor `unifi_countries` schema/state output shape.
- [x] Update tests for the new output attribute.
- [x] Update docs/examples and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Data source fetches beyond the API default page size.
- Data source no longer exposes output under `countries`.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- Changed `unifi_countries` output from `countries` to `items` (list of `{name, code}`).
- Updated schema tests to assert the new `items` attribute.
- Updated docs/examples to consume `data.unifi_countries.<name>.items`.
- Verified with `go test ./...`.
