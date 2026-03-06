# Todo

## Working Notes
- Add global countries lookup for firewall workflows:
- `unifi_countries` data source to expose country code/name pairs.

## Plan
- [ ] Implement `unifi_countries` data source.
- [ ] Register it in provider and extend schema/unit tests.
- [ ] Add documentation for `unifi_countries` and update README/docs index.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_countries` returns the full list of available countries with code and name.
- Provider/docs expose the new data source.
- Tests pass.

## Results
- In progress.
