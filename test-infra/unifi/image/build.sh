#!/bin/sh
# Build the UniFi OS Server test image.
#
#   UOS_VERSION        UniFi OS Server version (default below)
#   UOS_INSTALLER_URL  installer URL for the architecture being built; overrides
#                      the pinned pair for UOS_VERSION
#   UOS_IMAGE          tag to build (default unifi-os-server-test:dev)
#
# Ubiquiti's installer URLs carry a hash, so a version number alone cannot be
# resolved. The pinned pair below is the version this script was written for;
# to try another build, set UOS_VERSION together with UOS_INSTALLER_URL.
set -eu

: "${UOS_VERSION:=6.0.19}"
: "${UOS_IMAGE:=unifi-os-server-test:dev}"
arm64_url=https://fw-download.ubnt.com/data/unifi-os-server/08a8-linux-arm64-6.0.19-111cc795-a829-4f60-a6fb-ec4e15d5b54f.19-arm64
amd64_url=https://fw-download.ubnt.com/data/unifi-os-server/a6b0-linux-x64-6.0.19-4addcb6a-84d2-428c-84c3-d36a64740bdc.19-x64

here=$(cd "$(dirname "$0")" && pwd)
set -- --build-arg "UOS_VERSION=$UOS_VERSION" -t "$UOS_IMAGE"
if [ -n "${UOS_INSTALLER_URL:-}" ]; then
  set -- "$@" --build-arg "INSTALLER_URL_ARM64=$UOS_INSTALLER_URL" --build-arg "INSTALLER_URL_AMD64=$UOS_INSTALLER_URL"
elif [ "$UOS_VERSION" != 6.0.19 ]; then
  echo "UOS_VERSION=$UOS_VERSION has no pinned installer URL; set UOS_INSTALLER_URL as well" >&2
  exit 1
else
  set -- "$@" --build-arg "INSTALLER_URL_ARM64=$arm64_url" --build-arg "INSTALLER_URL_AMD64=$amd64_url"
fi
exec docker build "$@" "$here"
