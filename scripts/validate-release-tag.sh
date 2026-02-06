#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <tag>" >&2
  exit 1
fi

tag="$1"

if [[ ! "$tag" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)\+unifi\.([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "Tag must match vX.Y.Z+unifi.A.B.C, got: $tag" >&2
  exit 1
fi

if [[ -f "UNIFI_API_VERSION" ]]; then
  expected="$(tr -d ' \t\n\r' < UNIFI_API_VERSION)"
  if [[ -n "$expected" ]]; then
    if [[ ! "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+\+unifi\.${expected}$ ]]; then
      echo "Tag unifi version does not match UNIFI_API_VERSION (${expected})." >&2
      exit 1
    fi
  fi
fi

echo "Tag validated: ${tag}"
