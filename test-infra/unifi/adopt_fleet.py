#!/usr/bin/env python3
"""Adopts every device the unifi-emu fleet currently reports as pending,
into this console's one site, using nothing but the API key.

Unlike unifi-emu's own self-adoption (SIM_ADOPT), which needs a console
admin username/password, this uses the modern API-key-based
POST /v1/sites/{siteId}/devices (adoptDevice) endpoint. It needs
UNIFI_API_KEY in the environment (see README.md); the Makefile's
adopt-fleet target sources it from .env.
"""
import json
import os
import ssl
import sys
import time
import urllib.error
import urllib.request

API_URL = "https://localhost:11443/proxy/network/integrations"
TLS_CONTEXT = ssl._create_unverified_context()


def api(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(
        API_URL + path,
        data=data,
        method=method,
        headers={"X-API-Key": os.environ["UNIFI_API_KEY"], "Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, context=TLS_CONTEXT, timeout=15) as resp:
        return json.load(resp)


def wait_for(poll, description, timeout, interval):
    deadline = time.time() + timeout
    while time.time() < deadline:
        result = poll()
        if result is not None:
            return result
        time.sleep(interval)
    sys.exit(f"timed out after {timeout}s waiting for {description}")


def main():
    site_id = api("GET", "/v1/sites")["data"][0]["id"]

    # Wait for at least one pending device — but if the fleet is already
    # fully adopted, pending-devices will correctly stay empty forever, so
    # that alone can't be the loop condition. Check for existing adopted
    # devices on every empty poll and bail out cleanly the moment we see
    # one: that's "nothing to do", not "still waiting".
    deadline = time.time() + 60
    macs = []
    while time.time() < deadline:
        macs = [d["macAddress"] for d in api("GET", "/v1/pending-devices")["data"]]
        if macs:
            break
        existing = api("GET", f"/v1/sites/{site_id}/devices")["data"]
        if existing:
            print(f"site {site_id}: {len(existing)} device(s) already adopted, nothing pending — nothing to do")
            return
        time.sleep(2)
    else:
        sys.exit("timed out after 60s waiting for at least one pending device")

    print(f"site {site_id}: adopting {len(macs)} pending device(s): {', '.join(macs)}")

    for mac in macs:
        print(f"  adopting {mac}...")
        try:
            api("POST", f"/v1/sites/{site_id}/devices", {"macAddress": mac, "ignoreDeviceLimit": True})
        except urllib.error.HTTPError as err:
            sys.exit(f"adopt {mac} failed: HTTP {err.code} {err.read().decode(errors='replace')}")
        # Controllers reject adopt requests submitted in a burst; adopting
        # serially with a short gap is what unifi-emu's own self-adoption
        # does too (see its docs/USAGE.md).
        time.sleep(2)

    def all_online():
        devices = api("GET", f"/v1/sites/{site_id}/devices")["data"]
        states = {d["macAddress"]: d["state"] for d in devices}
        print(f"  states: {states}")
        adopted = {m: s for m, s in states.items() if m in macs}
        if len(adopted) == len(macs) and all(s == "ONLINE" for s in adopted.values()):
            return states
        return None

    wait_for(all_online, "every adopted device to reach ONLINE", timeout=120, interval=5)
    print("all devices connected")


if __name__ == "__main__":
    main()
