# Release Checklist

## Versioning

Releases are tagged using:

```
vX.Y.Z+unifi.A.B.C
```

- `X.Y.Z` is the provider version.
- `A.B.C` is the UniFi Network API version, stored in `UNIFI_API_VERSION`.

Use the helper to generate a tag:

```
./scripts/release-tag.sh 0.3.0
```

Then create and push the tag:

```
git tag -s "$(./scripts/release-tag.sh 0.3.0)" -m "Release 0.3.0"
git push origin "$(./scripts/release-tag.sh 0.3.0)"
```

The release workflow validates tag format and ensures `UNIFI_API_VERSION` matches.

## Terraform Registry (one-time setup)

1. Create the provider in Terraform Registry for the namespace `ilmax` and name `unifi`.
2. Upload your GPG public key to the provider settings.
3. Configure GitHub Actions secrets:
   - `GPG_PRIVATE_KEY`
   - `GPG_PASSPHRASE`

After setup, tagged releases will be ingested automatically.

## OpenTofu Registry (one-time setup)

1. Submit a provider request on the OpenTofu registry repo.
2. Provide the GitHub repo and your GPG public key.

After approval, tagged releases will be ingested automatically.

## Verify locally

- Run tests: `go test ./...`
- Build: `go build ./cmd/terraform-provider-unifi`

