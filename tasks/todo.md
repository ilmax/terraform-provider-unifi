# Todo

## Working Notes
- Implement `unifi_client` data source based on UniFi client API in SDK.

## Plan
- [x] Inspect SDK client endpoints and decide schema fields.
- [x] Implement data source and schema tests.
- [x] Add docs and update README/docs index.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `unifi_client` data source reads client details by ID.
- Docs and README updated.
- Tests pass.

## Results
- Added `unifi_client` data source with SDK variant handling and schema coverage.
- Documented the data source in docs and README.
- Verified with `go test ./...`.
