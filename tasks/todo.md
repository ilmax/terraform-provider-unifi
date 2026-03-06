# Todo

## Working Notes
- Add global countries lookup for firewall workflows:
- `unifi_countries` data source to expose country code/name pairs.

## Plan
- [x] Implement `unifi_countries` data source.
- [x] Register it in provider and extend schema/unit tests.
- [x] Add documentation for `unifi_countries` and update README/docs index.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- Added `unifi_countries` data source backed by `GET /v1/countries`.
- Registered the data source in the provider and added schema/unit tests.
- Added docs with an example map from country name to code for firewall workflows.
- Updated provider index and README data source listings.
- Verified with `go test ./...`.
