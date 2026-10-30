#!/bin/sh
# Builds unifi-emu with the `extra` inform-field patch as the local image
# unifi-emu:extra (override with IMAGE). Needs git and docker; nothing is
# installed on the host. Upstream: https://github.com/jamesbraid/unifi-emu
#
# The patch adds a per-device `extra` map to fleet files whose keys are merged
# into every inform payload, so the emulator can claim capabilities its model
# profiles lack (see ../README.md, "Patched emulator").
set -eu

VERSION="${UNIFI_EMU_VERSION:-v0.5.5}"
IMAGE="${IMAGE:-unifi-emu:extra}"
here="$(cd "$(dirname "$0")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

git clone -q --depth 1 --branch "$VERSION" https://github.com/jamesbraid/unifi-emu.git "$work/src"
git -C "$work/src" apply "$here/extra-fields.patch"
docker build -q --build-arg "VERSION=${VERSION#v}-extra" -t "$IMAGE" "$work/src"
echo "Built $IMAGE"
