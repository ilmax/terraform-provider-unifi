#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <provider_version (X.Y.Z)>" >&2
  exit 1
fi

provider_version="$1"
if [[ ! "$provider_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Provider version must be SemVer (X.Y.Z), got: $provider_version" >&2
  exit 1
fi

if [[ ! -f "UNIFI_API_VERSION" ]]; then
  echo "UNIFI_API_VERSION file not found" >&2
  exit 1
fi

unifi_version="$(tr -d ' \t\n\r' < UNIFI_API_VERSION)"
if [[ -z "$unifi_version" ]]; then
  echo "UNIFI_API_VERSION is empty" >&2
  exit 1
fi

if [[ ! "$unifi_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "UNIFI_API_VERSION must be in A.B.C format, got: $unifi_version" >&2
  exit 1
fi

tag="v${provider_version}+unifi.${unifi_version}"

if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  echo "Tag already exists: ${tag}" >&2
  exit 1
fi

echo "$tag"
