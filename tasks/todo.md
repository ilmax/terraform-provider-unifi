# Todo

## Working Notes
- Fix countries pagination:
- `unifi_countries` currently performs a single page request.
- Need to iterate with offset/limit until all countries are fetched.

## Plan
- [x] Implement paginated countries retrieval in `unifi_countries`.
- [x] Add regression tests for multi-page retrieval and stop conditions.
- [x] Update docs/results and verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Data source fetches beyond the API default page size.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- Updated `unifi_countries` to iterate through pages using `offset`/`limit` until all countries are collected.
- Added pagination safeguards (max pages and offset overflow checks).
- Added regression tests for multi-page retrieval, fallback stop condition, error propagation, and max-page guard.
- Updated countries data source documentation to clarify full pagination support.
- Verified with `go test ./...`.
