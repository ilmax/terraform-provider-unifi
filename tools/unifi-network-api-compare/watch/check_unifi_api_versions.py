#!/usr/bin/env python3
"""
check_unifi_api_versions.py — polling check for a new UniFi Network Application API.

Cheap signals (no decompile, no Docker, no 6.0 console):
  1. unifi-client-go `unifi-network-version` marker (the version its codegen targets)
  2. unifi-client-go newest spec under openapi/unifi-network/*.json
  3. public download catalog: newest unifi-os-server + network (self-host)
  4. community image: unihosted/unifi-os-server-docker tags

Fires an alert when the marker/spec goes above the known-good 10.6.106 (the provider's
current baseline) OR a 6.0 os-server / 11.x network deb appears OR a 6.0 image tag lands.

Exit codes: 0 = no change, 1 = change detected (alert), 2 = error.
Cron this (or the unifi-watch.sh wrapper). State is stored in state.json so repeated
runs only alert once per new version.
"""
import json, os, sys, re, urllib.request, ssl, datetime

HERE = os.path.dirname(os.path.abspath(__file__))
STATE = os.path.join(HERE, "state.json")
CTX = ssl.create_default_context(); CTX.check_hostname = False; CTX.verify_mode = ssl.CERT_NONE
BASELINE_VERSION = "10.6.106"   # the provider's current unifi-client-go spec

def get(url, timeout=20):
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "unifi-api-watch/1.0"})
        return urllib.request.urlopen(req, context=CTX, timeout=timeout).read().decode("utf-8", "replace")
    except Exception as e:
        return None

def vtuple(v):
    try:
        return tuple(int(x) for x in v.split(".")[:3])
    except Exception:
        return (0,)

def check():
    out = {"checked_at": datetime.datetime.now(datetime.timezone.utc).isoformat(), "alerts": [], "signals": {}}

    # 1) unifi-client-go version marker + newest spec
    marker = get("https://raw.githubusercontent.com/ilmax/unifi-client-go/main/unifi-network-version")
    marker = marker.strip() if marker else None
    out["signals"]["client_go_version_marker"] = marker

    tree = get("https://api.github.com/repos/ilmax/unifi-client-go/git/trees/main?recursive=1")
    newest_spec = None
    if tree:
        try:
            vers = [m.group(1) for t in json.loads(tree).get("tree", [])
                    for m in [re.search(r"unifi-network/(\d+\.\d+\.\d+)\.json$", t["path"])] if m]
            newest_spec = max(vers, key=vtuple) if vers else None
        except Exception:
            pass
    out["signals"]["client_go_newest_spec"] = newest_spec

    # 2) public catalogs
    try:
        products = json.loads(get("https://download.svc.ui.com/v1/downloads/products"))
        slugs = {p["id"]: p.get("slug") for p in products}
        os_ids = [k for k, v in slugs.items() if v == "unifi-os-server"]
        for pid in os_ids[:1]:
            ver = json.loads(get(f"https://download.svc.ui.com/v1/downloads/{pid}/versions?limit=1")).get("versions", [{}])[0]
            out["signals"]["catalog_newest_os_server"] = ver.get("version")
    except Exception as e:
        out["signals"]["catalog_newest_os_server"] = f"error: {e}"

    # 3) community image tags
    try:
        tags = [t["name"] for t in json.loads(get("https://hub.docker.com/v2/repositories/unihosted/unifi-os-server-docker/tags?page_size=100"))]
        out["signals"]["community_image_tags"] = sorted(set(tags))[:15]
        out["signals"]["community_image_has_6_0"] = any(t.startswith("6.") for t in tags)
    except Exception as e:
        out["signals"]["community_image_has_6_0"] = f"error: {e}"

    # ---- alert logic ----
    if marker and vtuple(marker) > vtuple(BASELINE_VERSION):
        out["alerts"].append(f"unifi-client-go version marker is {marker} (> {BASELINE_VERSION}) — new integrations spec likely; fetch + diff it")
    if newest_spec and vtuple(newest_spec) > vtuple(BASELINE_VERSION):
        out["alerts"].append(f"unifi-client-go has a new spec: {newest_spec} — diff against 10.6.106.json")
    os_server = out["signals"].get("catalog_newest_os_server")
    if os_server and vtuple(os_server) >= (6,):
        out["alerts"].append(f"public catalog now offers self-host os-server {os_server} — 6.0 is downloadable! Grab it, run the comparison tool.")
    if out["signals"].get("community_image_has_6_0") is True:
        out["alerts"].append("community unihosted/unifi-os-server-docker has a 6.0.x tag — build a 6.0 console, extract ace.jar, run compare_network_api.py")

    return out

def main():
    res = check()
    # dedupe: only alert for versions we haven't already alerted on
    prev = json.load(open(STATE)) if os.path.exists(STATE) else {"alerted_versions": []}
    seen = set(prev.get("alerted_versions", []))
    new_versions = [a for a in res["alerts"] if not any(s in a for s in seen)]
    # record which version/alert keys we've now seen
    for a in res["alerts"]:
        m = re.search(r"(\d+\.\d+\.\d+)", a)
        if m:
            seen.add(m.group(1))
    prev["alerted_versions"] = sorted(seen)
    prev["last_result"] = res
    json.dump(prev, open(STATE, "w"), indent=2)

    print(json.dumps(res, indent=2))
    if res["alerts"] and new_versions:
        print("\n*** ALERT *** new UniFi API surface detected:")
        for a in new_versions:
            print("  -", a)
        print("Next: run compare_network_api.py (next to this dir) against the new build/spec.")
        sys.exit(1)
    elif res["alerts"]:
        print(f"\n(i) {len(res['alerts'])} known alert(s) (already reported, not re-firing).")
        sys.exit(0)
    else:
        print(f"\nNo change. Marker={res['signals'].get('client_go_version_marker')} (baseline {BASELINE_VERSION}).")
        sys.exit(0)

if __name__ == "__main__":
    main()
