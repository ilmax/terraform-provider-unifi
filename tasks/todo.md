# Todo

## Working Notes
- Fix countries pagination:
- `unifi_countries` currently performs a single page request.
- Need to iterate with offset/limit until all countries are fetched.

## Plan
- [ ] Implement paginated countries retrieval in `unifi_countries`.
- [ ] Add regression tests for multi-page retrieval and stop conditions.
- [ ] Update docs/results and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Data source fetches beyond the API default page size.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- In progress.
