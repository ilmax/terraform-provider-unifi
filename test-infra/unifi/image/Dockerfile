# syntax=docker/dockerfile:1
#
# UniFi OS Server image for acceptance tests.
#
# Ubiquiti only ships UniFi OS Server as a self-extracting installer that wraps
# a systemd based OCI image and drives it with Podman. This recipe pulls that
# embedded image out of the installer and flattens it, so it can run under
# plain Docker. It is a test fixture, not a supported way to run UniFi OS.

FROM ubuntu:24.04 AS extract

ARG TARGETARCH
# Installer URLs come from build.sh, the single place that pins them (Ubiquiti's
# URLs carry a hash, so a version number alone cannot be resolved).
ARG INSTALLER_URL_ARM64
ARG INSTALLER_URL_AMD64

RUN apt-get update \
 && apt-get install -y --no-install-recommends binwalk ca-certificates curl jq p7zip-full unzip xz-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /work
RUN case "$TARGETARCH" in arm64) url="$INSTALLER_URL_ARM64" ;; amd64) url="$INSTALLER_URL_AMD64" ;; *) echo "unsupported arch $TARGETARCH" >&2; exit 1 ;; esac \
 && curl -fSL --retry 5 --retry-delay 2 -o installer "$url"

RUN binwalk --run-as=root -e installer >/dev/null 2>&1 || true; \
    find . -name image.tar -print -quit > image-path \
 && test -s image-path \
 && mkdir oci \
 && tar -xf "$(cat image-path)" -C oci \
 && rm -rf installer _installer.extracted

# Flatten the OCI layers into one root filesystem. Whiteouts are applied from
# the layer's own listing *before* extraction, so an opaque directory only
# hides lower layers and never the files the same layer just added.
COPY flatten-oci.sh /usr/local/bin/flatten-oci.sh
RUN chmod +x /usr/local/bin/flatten-oci.sh \
 && flatten-oci.sh /work/oci /rootfs \
 && m=$(jq -r '.manifests[0].digest' /work/oci/index.json | sed 's/^sha256://') \
 && c=$(jq -r '.config.digest' "/work/oci/blobs/sha256/$m" | sed 's/^sha256://') \
 && jq '.config' "/work/oci/blobs/sha256/$c" > /image-config.json

# ---------------------------------------------------------------------------
# Runtime image. Ubiquiti's own /root/uos-entrypoint.sh is the command; the
# installer normally supplies the variables below through Podman.
# ---------------------------------------------------------------------------
FROM scratch

COPY --from=extract /rootfs /

ARG TARGETARCH
ARG UOS_VERSION=6.0.19

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    FIRMWARE_PLATFORM=${TARGETARCH} \
    PRODUCT_NAME="UniFi OS Server" \
    APP_MODEL=UOSSERVER \
    APP_VERSION=${UOS_VERSION} \
    UOS_SERVER_VERSION=${UOS_VERSION} \
    UOS_UUID=00000000-0000-5000-8000-000000000001

STOPSIGNAL SIGRTMIN+3
CMD ["/root/uos-entrypoint.sh"]
