# Todo

## Working Notes
- Rework `unifi_countries.countries` to be a map keyed by country name.
- Desired usage: `data.unifi_countries.all.countries["China"].code`

## Plan
- [x] Change `unifi_countries.countries` from a list to a map keyed by country name.
- [x] Update tests and docs for the new access pattern.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- `data.unifi_countries.all.countries["China"].code` works.
- `countries` is exposed as a map keyed by country name.
- Tests pass.

## Results
- `unifi_countries.countries` is now a map keyed by country name.
- Value objects now expose `code`, enabling `countries["China"].code`.
- Updated tests and documentation for the new HCL usage pattern.
- Verified with `go test ./...`.
