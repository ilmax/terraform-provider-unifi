# Todo

## Working Notes
- Rework `unifi_countries.countries` to be a map keyed by country name.
- Desired usage: `data.unifi_countries.all.countries["China"].code`

## Plan
- [ ] Change `unifi_countries.countries` from a list to a map keyed by country name.
- [ ] Update tests and docs for the new access pattern.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- `data.unifi_countries.all.countries["China"].code` works.
- `countries` is exposed as a map keyed by country name.
- Tests pass.

## Results
- In progress.
