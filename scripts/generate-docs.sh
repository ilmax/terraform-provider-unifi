#!/bin/sh
# Regenerates docs/ (index, resources, data sources, actions) from the
# provider's schema with tfplugindocs.
#
# tfplugindocs normally builds the provider from the repository root, but the
# provider's main package lives in cmd/terraform-provider-unifi, so this builds
# it, exports its schema through the Terraform CLI (a dev override, no
# registry access), and hands the schema to tfplugindocs.
#
#   DOCS_SCHEMA_OUT  also copy the exported schema JSON to this path (and skip
#                    generation), for tooling that needs the schema.
#
# Needs Go and the Terraform CLI. tfplugindocs is run at a pinned version, so
# the output does not change when a new release appears. Run it with `make docs`;
# CI runs it and fails when docs/ is out of date.
set -eu

TFPLUGINDOCS_VERSION=v0.24.0

root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

(cd "$root" && go build -o "$tmp/bin/terraform-provider-unifi" ./cmd/terraform-provider-unifi)

cat > "$tmp/terraformrc" <<EOF
provider_installation {
  dev_overrides { "registry.terraform.io/hashicorp/unifi" = "$tmp/bin" }
  direct {}
}
EOF

mkdir "$tmp/ws"
cat > "$tmp/ws/main.tf" <<'EOF'
terraform {
  required_providers { unifi = { source = "hashicorp/unifi" } }
}
EOF
(cd "$tmp/ws" && TF_CLI_CONFIG_FILE="$tmp/terraformrc" terraform providers schema -json > "$tmp/schema.json")

if [ -n "${DOCS_SCHEMA_OUT:-}" ]; then
  cp "$tmp/schema.json" "$DOCS_SCHEMA_OUT"
  exit 0
fi

cd "$root"
go run "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$TFPLUGINDOCS_VERSION" generate \
  --provider-name unifi \
  --rendered-provider-name UniFi \
  --providers-schema "$tmp/schema.json"
