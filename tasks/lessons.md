# Lessons

- Before assuming provider drift bugs, verify the Terraform config actually sets the attributes being compared; locals not wired into resources won't produce diffs.
- If state does not reflect upstream data, review Read mapping and schema optional/computed flags before assuming plan is correct.
- Plugin Framework defaults require `Computed=true`; prefer plan modifiers for optional defaults.
- Plan modifiers cannot inject non-computed defaults without triggering invalid plan errors.
- When consuming list endpoints, do not assume one request returns all results; verify pagination metadata (`offset`, `limit`, `totalCount`) and add regression tests for multi-page reads.
- Keep Terraform resource/data source naming discoverable by aligning file names with exposed type names (for example `unifi_firewall_rule` -> `resource_firewall_rule.go`), and verify this during reviews.
- For Terraform UX-heavy data sources, shape state for the intended HCL access pattern first; if users need keyed lookups, prefer maps over lists plus comprehensions.
- When adding provider-level defaults like `site_id`, document the precedence once in repo instructions and add schema/helper regression tests so new resources and data sources cannot drift from the pattern.
- Before a multi-step task, confirm the repository is the right one (`git remote -v`, the `go.mod` module path, the SDK it imports). A Network 11 test environment was built in the wrong repository because the working directory was assumed to be the project; a request that names a dependency the code does not use is the signal to stop and ask.
- A test helper must do what its name says. `hasAdoptedGateway` returned true for any adopted device, so it only worked while the emulated fleet always contained a gateway; the tests it guarded failed as soon as the fleet changed.
- Do not assert counts or states that belong to one console build (for example "exactly 7 zones"). Assert the property that matters (at least the default zones) so a console upgrade does not break the test.
- An emulated device can change console-wide behavior, not only its own: adopting the emulator's older-capability gateway makes UniFi Network 11 drop the site's firewall zones. After changing the fleet, check the console-level features the tests rely on, not just that the devices come online.
- UniFi Network 11 no longer keeps its configuration in MongoDB at runtime; scripts that write to Mongo (`enable_zbf.sh`) cannot work there. Check where a version keeps its data before reusing a database-level workaround.
