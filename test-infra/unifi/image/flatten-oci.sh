#!/bin/sh
# Usage: flatten-oci.sh <oci-layout-dir> <rootfs-dir>
# Applies every layer of the first manifest, in order, onto <rootfs-dir>.
set -eu
oci=$1 root=$2
blob() { echo "$oci/blobs/sha256/${1#sha256:}"; }
manifest=$(blob "$(jq -r '.manifests[0].digest' "$oci/index.json")")
mkdir -p "$root"
for digest in $(jq -r '.layers[].digest' "$manifest"); do
  layer=$(blob "$digest")
  tar -tf "$layer" | while IFS= read -r entry; do
    name=${entry##*/}
    dir=${entry%/*}; [ "$dir" = "$entry" ] && dir=.
    case "$name" in
      .wh..wh..opq) find "$root/$dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} + 2>/dev/null || true ;;
      .wh.*)        rm -rf "$root/$dir/${name#.wh.}" ;;
    esac
  done
  tar -xf "$layer" -C "$root" --exclude='.wh.*' --exclude='*/.wh.*'
done
