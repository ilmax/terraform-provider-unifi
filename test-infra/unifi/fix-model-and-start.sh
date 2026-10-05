#!/bin/bash
# hieutq/unifi-os-server's own /entrypoint.sh writes /usr/lib/platform and
# /usr/lib/version, but never /usr/lib/product_name or /usr/lib/app_model.
# unifi-core shells out to /sbin/ubnt-tools, which reads those two files to
# report `board.shortname`/`board.sysid` — with them missing, ubnt-tools
# reports an empty shortname and unifi-core crash-loops with
# `Error: Unsupported console model: ""`.
#
# The official installer sets these via env vars (APP_MODEL=UOSSERVER,
# PRODUCT_NAME="UniFi OS Server") and writes them to these same paths
# inside its own image; this wrapper does the same before handing off to
# the image's real entrypoint.
set -e

echo -n "${APP_MODEL:-UOSSERVER}" > /usr/lib/app_model
echo -n "${PRODUCT_NAME:-UniFi OS Server}" > /usr/lib/product_name

# Pin the Network application to a specific version, if requested. The
# bundled default (whatever version was baked into this image at build
# time) is otherwise unrelated to the UOS_SERVER image tag/version.
#
# `uos runnable install` manages a running system (it talks to systemd and
# the app's own services), so it can't run here — systemd hasn't started
# yet, this whole script runs before `exec /sbin/init` below. Instead,
# background a watcher that waits for unifi.service to come up, checks the
# installed version, and installs the requested one if it doesn't match.
# Backgrounding survives the `exec` below: forked children aren't affected
# by their parent's process image being replaced.
if [ -n "${NETWORK_VERSION:-}" ]; then
    (
        for _ in $(seq 1 120); do
            [ "$(systemctl is-active unifi.service 2>/dev/null)" = "active" ] && break
            sleep 5
        done

        CURRENT="$(uos runnable current-version unifi 2>/dev/null || true)"
        case "$CURRENT" in
            "${NETWORK_VERSION}"*)
                echo "[fix-model-and-start] Network app already at ${CURRENT}, matching requested ${NETWORK_VERSION}"
                ;;
            *)
                echo "[fix-model-and-start] Installing Network app ${NETWORK_VERSION} (current: ${CURRENT:-unknown})..."
                uos runnable install unifi --version "${NETWORK_VERSION}" --progress
                echo "[fix-model-and-start] Network app now at $(uos runnable current-version unifi 2>/dev/null || true)"
                ;;
        esac
    ) &
fi

exec /entrypoint.sh "$@"
