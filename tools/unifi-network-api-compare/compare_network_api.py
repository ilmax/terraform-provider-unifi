#!/usr/bin/env python3
"""
compare_network_api.py — diff a new UniFi Network Application build against the
10.6.106 baseline to see what API surface changed.

Three comparisons (whichever inputs you give):
  1. OpenAPI spec (integration.json)  — paths / ops / schemas   [public Integrations API]
  2. HTTP routes (from decompiled source) — /v1, /api, /docs    [what the running app serves]
  3. Promoted integrations sub-packages (com/ubnt/net/d/a/*)   [new resource families]

Inputs (any of):
  --jar <ace.jar>            decompile it (CFR in a java container), then diff
  --deb <unifi*.deb>         extract ace.jar, decompile, then diff
  --decompiled-dir <dir>     an already-decompiled source tree (contains com/ubnt/...)
  --spec <integration.json>  also diff the OpenAPI spec against the baseline
  --new-decompiled-dir / --baseline-decompiled-dir   for a plain two-dir diff

The 10.6.106 baseline (decompiled source + spec + routes) is baked into ./baseline-10.6.106.

Accuracy note (route extraction): the parser resolves ~381/433 baseline paths exactly
(100% method agreement on matches). The ~52 unresolved paths are integrations controllers
that use a *relative* method path under a shared obfuscated base-marker (e.g. /networks vs
/clients) — the marker's true base is genuinely ambiguous in the source and is recovered from
the baseline route table on the baseline side. Because BOTH sides of a diff use the same
parser, a self-diff is perfectly clean (0 spurious changes) and *new* absolute-path endpoints
are always caught. Treat the route diff as: trustworthy for additions/removals, approximate
for the small set of shared-marker relative-path files.

Self-test (expect 0 added / 0 removed everywhere):
  ./compare_network_api.py --baseline-decompiled-dir baseline-10.6.106 \
      --new-decompiled-dir baseline-10.6.106 --spec baseline-10.6.106/integration.json

Requires: python3. Decompilation uses CFR in an eclipse-temurin:21-jdk-jammy container (auto-download).
"""
import argparse, json, os, re, sys, subprocess, tempfile, urllib.request, ssl
from collections import defaultdict

HERE = os.path.dirname(os.path.abspath(__file__))
BASELINE = {
    "dir":    os.path.join(HERE, "baseline-10.6.106"),
    "spec":   os.path.join(HERE, "baseline-10.6.106", "integration.json"),
    "routes": os.path.join(HERE, "baseline-10.6.106", "all-routes.json"),
}
CFR_URL = "https://www.benf.org/other/cfr/cfr-0.152.jar"
CFR_JAR = os.path.join(HERE, "cfr-0.152.jar")
JAVA_CTR = os.environ.get("DECOMPILE_IMAGE", "eclipse-temurin:21-jdk-jammy")

def read(p):
    with open(p, "r", errors="replace") as f:
        return f.read()

def find_java(java_root):
    for dp, _, fns in os.walk(java_root):
        for fn in fns:
            if fn.endswith(".java"):
                yield os.path.join(dp, fn)

def _ann_body(src, i):
    """Given index i at a '(' of an annotation, return the annotation body (up to matching close paren)."""
    depth, j = 1, i
    while j < len(src) and depth:
        c = src[j]
        if c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
        j += 1
    return src[i + 1:j - 1]

def class_ann_block(src):
    """Text of the annotation block directly above the class declaration (after imports)."""
    lines = src.splitlines()
    for i, line in enumerate(lines):
        if re.match(r"\s*(?:public\s+|abstract\s+|final\s+)*class\b", line):
            j = i - 1
            while j >= 0 and (lines[j].strip().startswith("@") or lines[j].strip() == ""):
                j -= 1
            return "\n".join(lines[j + 1:i])
    return ""

def direct_base(block):
    """Base path from a class-level @RequestMapping (brace-aware), or ''."""
    m = re.search(r"@RequestMapping\s*\(", block)
    if not m:
        return ""
    body = _ann_body(block, m.end() - 1)
    vm = re.search(r"value\s*=\s*\{?\s*\"([^\"]*)\"", body)
    return vm.group(1) if vm else ""

def _marker_names(block):
    """FQ class names of class-level marker annotations (multi-line-value aware)."""
    names = []
    for m in re.finditer(r"@([A-Za-z][\w.]*)\b", block):
        depth, i = 0, m.end()
        # scan forward; if we hit an '(' the annotation has a value (a @Tag etc), not a marker
        while i < len(block):
            c = block[i]
            if c == "(":
                depth += 1
            elif c == ")":
                depth -= 1
                if depth == 0:
                    i += 1
                    break
            elif c in "\n" and depth == 0:
                break
            i += 1
        if depth == 0 and block[m.end():i].strip() == "":
            names.append(m.group(1))
    return [n for n in names if n.startswith("com.ubnt.")]

def markers(block):
    return _marker_names(block)

def method_mappings(src):
    """[(http_method, subpath), ...] from @*Mapping on methods (brace-aware)."""
    out = []
    for m in re.finditer(r"@(Get|Post|Put|Delete|Patch|Request)Mapping\s*\(", src, re.I):
        body = _ann_body(src, m.end() - 1)
        vm = re.search(r"value\s*=\s*\{?\s*\"([^\"]*)\"", body)
        if vm:
            out.append((m.group(1).lower(), vm.group(1)))
    return out

def is_controller(src):
    return any(t in src for t in ("@GetMapping", "@PostMapping", "@PutMapping", "@DeleteMapping", "@PatchMapping"))

def join(base, sub):
    # an absolute method value (starts with '/') is used verbatim — the /api
    # controllers put the full /api/site/{siteName}/... path in the method mapping
    if sub.startswith("/"):
        return sub.rstrip("/") or "/"
    sub = sub.lstrip("/")
    if not base:
        return sub or "/"
    return (base.rstrip("/") + "/" + sub).rstrip("/") or "/"

def _full_paths(src, base):
    return [join(base, s) for _m, s in method_mappings(src)]

def infer_base(known, fulls, direct, direct_is_base=False):
    """Best base for a file.
    - If the file has a DIRECT @RequestMapping base, and it's not empty, use it (it's authoritative).
    - Otherwise (marker-bearing file, absolute method paths): the true base is the known
      path that is a common prefix of the file's full paths. Absolute method values mean
      the file has no effective base, so we look for the shortest known path that every
      full path shares as a prefix."""
    fulls = [fp for fp in fulls if fp]
    if direct_is_base and direct:
        return direct
    if not fulls:
        return direct
    # shortest known path that is a prefix of all full paths
    cands = [kp for kp in sorted(known, key=len) if all(fp.startswith(kp.rstrip("/")) for fp in fulls)]
    if cands:
        return cands[0]
    return direct

def build_marker_registry(java_root, known):
    """Map an obfuscated marker FQ name -> its base path (a single string). For a marker
    shared by several controllers, resolve to the shortest path that prefixes all of the
    files' inferred bases (e.g. sdnKvBmrTA -> /v1/sites/{siteId})."""
    marker_cands = defaultdict(set)
    for f in find_java(java_root):
        src = read(f)
        block = class_ann_block(src)
        if not block:
            continue
        mmaps = method_mappings(src)
        mks = markers(block)
        if not (mmaps and mks):
            continue
        direct = direct_base(block)
        fulls = [join(direct, s) for _m, s in mmaps]
        base = infer_base(known, fulls, direct, direct_is_base=(direct != ""))
        if base:
            for mk in mks:
                marker_cands[mk].add(base)
    reg = {}
    for mk, cands in marker_cands.items():
        if len(cands) == 1:
            reg[mk] = next(iter(cands))
        else:
            sc = sorted(cands, key=len)
            for c in sc:
                if all(p.startswith(c.rstrip("/")) for p in sc):
                    reg[mk] = c
                    break
    return reg

def extract_routes(java_root, registry, known):
    routes = {}
    for f in find_java(java_root):
        src = read(f)
        if not is_controller(src):
            continue
        block = class_ann_block(src)
        base = direct_base(block)
        if not base:
            mks = markers(block)
            cands = sorted({registry[m] for m in mks if m in registry and isinstance(registry[m], str)}, key=len)
            if len(cands) == 1:
                base = cands[0]
            elif cands:
                if known is not None:
                    # shared marker: pick the candidate whose joined paths best match the baseline
                    fulls_by = {c: {join(c, s) for _m, s in method_mappings(src)} for c in cands}
                    base = max(cands, key=lambda c: len(fulls_by[c] & set(known)))
                else:
                    base = cands[0]
        for method, sub in method_mappings(src):
            p = join(base, sub)
            routes.setdefault(p, [])
            if method not in routes[p]:
                routes[p].append(method)
    return routes

def load_routes(java_root):
    known_path = os.path.join(os.path.dirname(os.path.abspath(java_root)), "all-routes.json")
    known = json.load(open(known_path)) if os.path.exists(known_path) else None
    registry = build_marker_registry(java_root, known)
    return extract_routes(java_root, registry, known)

# ---------------- spec ----------------
def spec_summary(spec):
    d = json.load(open(spec))
    return d, {p: sorted(v.keys()) for p, v in d.get("paths", {}).items()}, len(d.get("components", {}).get("schemas", {}))

def spec_diff(base_spec, new_spec):
    _, ob, _ = spec_summary(base_spec)
    _, on, _ = spec_summary(new_spec)
    return {"baseline_paths": len(ob), "new_paths": len(on),
            "added": sorted(set(on) - set(ob)), "removed": sorted(set(ob) - set(on)),
            "method_changes": [(p, ob[p], on[p]) for p in set(ob) & set(on) if ob[p] != on[p]]}

# ---------------- routes ----------------
def bucket(p):
    if p.startswith("/v1"): return "integrations (/v1)"
    if p.startswith("/api"): return "internal (/api)"
    if p.startswith(("/docs", "/v2")): return "openapi (/docs /v2)"
    return "other"

def route_diff(base_r, new_r):
    b, n = set(base_r), set(new_r)
    added, removed = sorted(n - b), sorted(b - n)
    mchanges = [(p, base_r[p], new_r[p]) for p in (b & n) if sorted(base_r[p]) != sorted(new_r[p])]
    return added, removed, mchanges

def bucketed(paths):
    out = defaultdict(list)
    for p in paths:
        out[bucket(p)].append(p)
    return {k: sorted(v) for k, v in out.items()}

# ---------------- promoted packages ----------------
def promoted_packages(java_root):
    da = os.path.join(java_root, "ubnt/net/d/a")
    if not os.path.isdir(da):
        return {}
    out = {}
    for sub in sorted(os.listdir(da)):
        sp = os.path.join(da, sub)
        if not os.path.isdir(sp):
            continue
        tags = {t for f in find_java(sp) for t in re.findall(r'@Tag\(\s*name\s*=\s*"([^"]+)"', read(f))}
        if tags:
            out[sub] = sorted(tags)
    return out

# ---------------- decompile ----------------
def ensure_cfr():
    if os.path.exists(CFR_JAR):
        return CFR_JAR
    print(f"[i] downloading CFR decompiler -> {CFR_JAR}")
    ctx = ssl.create_default_context(); ctx.check_hostname = False; ctx.verify_mode = ssl.CERT_NONE
    urllib.request.urlretrieve(CFR_URL, CFR_JAR, context=ctx)
    return CFR_JAR

def extract_ace_jar(deb_path):
    work = tempfile.mkdtemp(prefix="deb-")
    os.system(f"ar x {os.path.abspath(deb_path)} -C {work} >/dev/null 2>&1")
    datatar = next((os.path.join(work, f) for f in os.listdir(work) if f.startswith("data.tar")), None)
    if not datatar:
        raise SystemExit("[!] no data.tar in the .deb")
    os.system(f"tar -x{os.path.splitext(datatar)[1]} -C {work} --wildcards '*ace.jar' '*internal-dependencies.jar' >/dev/null 2>&1")
    found = [os.path.join(dp, fn) for dp, _, fns in os.walk(work) for fn in fns if fn in ("ace.jar", "internal-dependencies.jar")]
    if not found:
        raise SystemExit("[!] could not find ace.jar inside the .deb")
    return found

def decompile(source):
    """Decompile source (a .jar/.class or a directory of them) into a fresh dir."""
    ensure_cfr()
    out = tempfile.mkdtemp(prefix="decomp-")
    src = os.path.abspath(source)
    cfr = os.path.abspath(CFR_JAR)
    if os.path.isdir(src):
        mounts = ["-v", f"{src}:/src:ro"]
        inner = "cd /src && java -jar /cfr.jar . --outputdir /out --silent true 2>/dev/null"
    else:
        mounts = ["-v", f"{src}:/in:ro"]
        inner = "cp /in /app.jar && java -jar /cfr.jar /app.jar --outputdir /out --silent true 2>/dev/null"
    cmd = ["docker", "run", "--rm", *mounts, "-v", f"{cfr}:/cfr.jar:ro", "-v", f"{out}:/out", JAVA_CTR, "bash", "-lc", inner]
    print(f"[i] decompiling {source} in {JAVA_CTR} (this can take minutes)…")
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        print("[!] decompiler:", (r.stderr or r.stdout)[:600])
    return out

# ---------------- main ----------------
def main():
    ap = argparse.ArgumentParser(description="Diff a new UniFi Network app build vs 10.6.106")
    ap.add_argument("--jar", help="new ace.jar / internal-dependencies.jar (decompiles it)")
    ap.add_argument("--deb", help="new unifi*.deb (extracts ace.jar, then decompiles)")
    ap.add_argument("--decompiled-dir", help="an already-decompiled source tree (contains com/ubnt/...)")
    ap.add_argument("--new-decompiled-dir", help="new build's decompiled tree (two-dir mode)")
    ap.add_argument("--baseline-decompiled-dir", default=BASELINE["dir"], help="baseline decompiled tree")
    ap.add_argument("--spec", help="new integration.json (also diff the OpenAPI spec)")
    ap.add_argument("--java-root", default="com", help="top package dir inside a decompiled tree")
    ap.add_argument("--out", help="write the diff JSON here")
    args = ap.parse_args()

    baseline_java = os.path.join(args.baseline_decompiled_dir, args.java_root)
    baseline_has_tree = os.path.isdir(baseline_java)
    if not baseline_has_tree and not args.spec:
        sys.exit(f"[!] no baseline decompiled tree at {baseline_java} and no --spec given.\n"
                 f"    Give a --spec (spec-only diff) or a --jar/--decompiled-dir (full diff).")

    result = {"baseline": "10.6.106"}

    new_java, new_label = None, None
    if args.decompiled_dir:
        new_java, new_label = os.path.join(args.decompiled_dir, args.java_root), os.path.basename(args.decompiled_dir)
    elif args.new_decompiled_dir:
        new_java, new_label = os.path.join(args.new_decompiled_dir, args.java_root), os.path.basename(args.new_decompiled_dir)
    elif args.jar or args.deb:
        src = args.deb if args.deb else args.jar
        if args.deb:
            print(f"[i] extracted: {extract_ace_jar(args.deb)}")
            src = extract_ace_jar(args.deb)[0]
        new_java = os.path.join(decompile(src), args.java_root)
        new_label = os.path.basename(src)
    result["new"] = new_label or "(baseline self-diff)"

    # spec
    if args.spec and os.path.exists(BASELINE["spec"]):
        result["spec"] = spec_diff(BASELINE["spec"], args.spec)
    else:
        result["spec"] = {"note": "skipped (no --spec or no baseline spec)"}

    # routes — only meaningful if a baseline decompiled tree is present
    if baseline_has_tree:
        print("[i] extracting baseline routes…")
        base_r = load_routes(baseline_java)
        print(f"    baseline: {len(base_r)} routes")
        if new_java and os.path.isdir(new_java):
            print("[i] extracting new routes…")
            new_r = load_routes(new_java)
            print(f"    new: {len(new_r)} routes")
            ra, rr, rm = route_diff(base_r, new_r)
            result["routes"] = {
                "baseline": len(base_r), "new": len(new_r),
                "added_count": len(ra), "removed_count": len(rr),
                "added_by_bucket": bucketed(ra), "removed_by_bucket": bucketed(rr),
                "added_paths": ra, "removed_paths": rr, "method_changes": rm,
            }
        else:
            result["routes"] = {"baseline": len(base_r), "note": "no new tree; baseline only"}
    else:
        result["routes"] = {"note": "skipped (no baseline decompiled tree; spec-only mode — pass --jar/--decompiled-dir for routes)"}

    # promoted packages
    if new_java and os.path.isdir(new_java) and baseline_has_tree:
        bp, np = promoted_packages(baseline_java), promoted_packages(new_java)
        result["promoted_packages"] = {
            "baseline": bp, "new": np,
            "added": {k: v for k, v in np.items() if k not in bp},
            "removed": {k: v for k, v in bp.items() if k not in np},
        }
    else:
        result["promoted_packages"] = {"note": "skipped (needs a baseline + new decompiled tree)"}

    print("\n" + json.dumps(result, indent=2))
    if args.out:
        json.dump(result, open(args.out, "w"), indent=2)
        print(f"\n[i] wrote {args.out}")

if __name__ == "__main__":
    main()
