# Lessons

- Before assuming provider drift bugs, verify the Terraform config actually sets the attributes being compared; locals not wired into resources won't produce diffs.
- If state does not reflect upstream data, review Read mapping and schema optional/computed flags before assuming plan is correct.
- Plugin Framework defaults require `Computed=true`; prefer plan modifiers for optional defaults.
- Plan modifiers cannot inject non-computed defaults without triggering invalid plan errors.
- When consuming list endpoints, do not assume one request returns all results; verify pagination metadata (`offset`, `limit`, `totalCount`) and add regression tests for multi-page reads.
- Keep Terraform resource/data source naming discoverable by aligning file names with exposed type names (for example `unifi_firewall_rule` -> `resource_firewall_rule.go`), and verify this during reviews.
- For Terraform UX-heavy data sources, shape state for the intended HCL access pattern first; if users need keyed lookups, prefer maps over lists plus comprehensions.
