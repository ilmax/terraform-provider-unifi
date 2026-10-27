#!/bin/sh
# Install a UniFi Network application version into a running UniFi OS Server
# container. Network 11 is not offered by the OS's own updater yet, so the
# package URL is looked up in Ubiquiti's firmware API and passed straight to
# `uos runnable install`.
#
#   NETWORK_VERSION  version to install (default below), e.g. 11.0.97
#   NETWORK_URL      package URL; skips the lookup
#   NETWORK_CHANNEL  firmware channel for the lookup (default release)
#   UOS_CONTAINER          container name (default uos-test)
#
# Safe to rerun: it does nothing when that version is already installed.
set -eu

: "${NETWORK_VERSION:=$(tr -d ' \t\n\r' < "$(dirname "$0")/../../UNIFI_API_VERSION")}"
: "${NETWORK_CHANNEL:=release}"
: "${UOS_CONTAINER:=uos-test}"

uos() { docker exec "$UOS_CONTAINER" uos "$@"; }

current=$(uos runnable current-version unifi 2>/dev/null || true)
case $current in
  "$NETWORK_VERSION"-*) echo "UniFi Network $current already installed"; exit 0 ;;
esac

url=${NETWORK_URL:-}
if [ -z "$url" ]; then
  arch=$(docker exec "$UOS_CONTAINER" dpkg --print-architecture)
  api="https://fw-update.ubnt.com/api/firmware?filter=eq~~product~~unifi&filter=eq~~channel~~$NETWORK_CHANNEL&filter=eq~~platform~~uos-deb13-$arch&sort=-created&limit=100"
  url=$(curl -fsS "$api" | python3 -I -c '
import json, sys
want = "v" + sys.argv[1] + "-"
for fw in json.load(sys.stdin).get("_embedded", {}).get("firmware", []):
    if fw["version"].startswith(want):
        print(fw["_links"]["data"]["href"]); break
' "$NETWORK_VERSION")
  [ -n "$url" ] || { echo "no UniFi Network $NETWORK_VERSION package for $arch on channel $NETWORK_CHANNEL; set NETWORK_URL" >&2; exit 1; }
fi

echo "Installing UniFi Network $NETWORK_VERSION from $url"
uos runnable install unifi --direct-url "$url" >/dev/null

# Wait until the application itself is up.
i=0
until [ "$(docker exec "$UOS_CONTAINER" systemctl is-active unifi 2>/dev/null)" = active ]; do
  i=$((i + 1)); [ $i -le 90 ] || { echo "unifi service did not become active" >&2; exit 1; }
  sleep 5
done
echo "UniFi Network $(uos runnable current-version unifi) is running"
