# UniFi Network API Comparison + Watch

Tooling to detect **what changed in the UniFi Network Application API** when a new
build ships (e.g. 10.6.106 → 11.x), and to *watch* for new public API surface
automatically. See [`../../docs/api-versioning-and-promotion.md`](../../docs/api-versioning-and-promotion.md)
for the background (Integrations API vs Internal API, and how Ubiquiti promotes
resources).

```
tools/unifi-network-api-compare/
├── compare_network_api.py        # diff a new build vs the 10.6.106 baseline
├── baseline-10.6.106/            # the 10.6.106 baseline (spec + route table)
│   ├── integration.json          #   44-path OpenAPI Integrations API spec
│   └── all-routes.json           #   433-route surface (path -> methods)
└── watch/                        # the low-cost polling watch
    ├── check_unifi_api_versions.py
    └── unifi-watch.sh            # cron wrapper
```

## Why this exists

The public **Integrations API** (`/v1`, API-key auth) is a **hand-curated** subset
of the ~433 HTTP routes the Network app actually serves. Ubiquiti promotes a handful
of resources per release and leaves the rest on the internal `/api` (session-cookie)
surface. To know *what* Ubiquiti promoted in a new release — and whether it's worth
a new client version + resource work — you diff the new build against the baseline on
three independent signals:

1. **OpenAPI spec** (`integration.json`) — the documented integrations surface
   (44 paths in 10.6.106). The signal `unifi-client-go` and the provider track. *Exact.*
2. **HTTP routes** (from decompiled source) — the full 433-route surface. Catches
   endpoints the spec doesn't list.
3. **Promoted integrations packages** (`com/ubnt/net/d/a/<x>`) — *new resource
   families* in the integrations code (a 19th sub-package = a new promoted family).

## The baseline

`baseline-10.6.106/` holds the 10.6.106 **OpenAPI spec** and the **route table**.
This is what the provider currently tracks, so it's the "known-good" reference.

> Note: the baseline is the *small* artifacts (spec + route table). The decompiled
> Java **source tree** is intentionally *not* kept in the repo — `compare_network_api.py`
> decompiles a jar on demand (see below), so no 60MB of generated code lives here.

## Usage

```bash
# Spec-only diff (fast, no decompile) — the primary signal:
./compare_network_api.py --spec /path/to/NEW-integration.json

# Full 3-signal diff — give it a new ace.jar or internal-dependencies.jar:
./compare_network_api.py --jar /path/to/NEW-ace.jar          # auto-decompiles
./compare_network_api.py --deb /path/to/unifi_11.0.81.deb    # extracts ace.jar first
./compare_network_api.py --decompiled-dir /path/to/decompiled  # pre-decompiled tree

# Write the full diff as JSON:
./compare_network_api.py --spec /path/to/NEW-integration.json --out diff-11.0.81.json
```

The decompile path auto-downloads the CFR decompiler (`cfr-0.152.jar`) and runs it
in an `eclipse-temurin:21-jdk-jammy` container (override with `DECOMPILE_IMAGE=…`).

### Output

JSON with `spec` (added/removed/changed paths + ops), `routes` (added/removed by
`/v1` vs `/api` bucket + method-level changes), and `promoted_packages` (new/removed
`com/ubnt/net/d/a/*` families with their `@Tag` names). A **new integrations family**
shows up as an entry in `promoted_packages.added` — that's your "new resource" alert.

### Getting a new build to analyze

11.0.81 isn't publicly downloadable (see the docs page). Realistic sources, in order:
1. **`unifi-client-go`** regenerating a 11.x spec — diff the *spec* only (`--spec`);
   no decompile needed. (This is what the watch polls for.)
2. **UOS 6.0 community Docker image** — run it, `docker cp` the `ace.jar`, full diff.
3. **A running 11.x console** — `curl` `/integration/api-docs` with an API key.

## The watch

`watch/check_unifi_api_versions.py` polls cheaply (no decompile, no Docker) for a new
public API surface. **Primary signal:** `unifi-client-go`'s `unifi-network-version`
marker + newest `openapi/unifi-network/*.json` spec on `main` — fires when either
goes above 10.6.106. Best-effort bonus signals: the public `download.svc.ui.com`
catalog for a 6.0 os-server / 11.x network `.deb`, and the community
`unihosted/unifi-os-server-docker` image for a 6.0 tag (these degrade gracefully if
rate-limited/unreachable). State is stored in `watch/state.json` so it only fires once
per new version.

Install as a cron job (adjust to your absolute path):
```cron
0 */6 * * * /path/to/tools/unifi-network-api-compare/watch/unifi-watch.sh >> /path/to/tools/unifi-network-api-compare/watch/watch.log 2>&1
```

When it fires, run the comparison tool against the new build/spec it names.

## Caveats

- The **route diff** is ~88% path-accurate against the baseline (381/433 exact; 100%
  method agreement on matches). The ~52 unresolved paths are integrations controllers
  using a *relative* path under a shared obfuscated base-marker, resolved from the
  baseline route table on the baseline side. Both diff sides use the same parser, so a
  self-diff is clean and *new absolute-path* endpoints are always caught. **Treat the
  OpenAPI-spec diff (signal #1) as the primary, exact thing to trust.**
- Full decompilation of a jar takes several minutes and a few GB of disk.
- The watch's Docker Hub / Ubiquiti-catalog signals can be rate-limited; the GitHub
  `unifi-client-go` signal is the reliable one.
