TEST         ?= ./...
TESTARGS     ?=
TEST_COUNT   ?= 1
TEST_TIMEOUT ?= 10m

ACCTEST_PKG   ?= ./internal/acceptance/...
CONSOLE_DIR   := test-infra/unifi
CONSOLE_ENV   := $(CONSOLE_DIR)/.env

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -count $(TEST_COUNT) $(TEST) $(TESTARGS)

# testacc runs the real acceptance suite in internal/acceptance against a
# live UniFi OS Server console (see test-infra/unifi/README.md — `make
# console-up` starts one locally). Credentials come from
# test-infra/unifi/.env; UNIFI_SITE_ID is resolved dynamically because the
# real API needs the site's UUID, not the console's "default" display name,
# and that UUID isn't known until the console exists.
.PHONY: testacc
testacc:
	@test -f $(CONSOLE_ENV) || { echo "Missing $(CONSOLE_ENV) — see $(CONSOLE_DIR)/README.md to create one."; exit 1; }
	@set -a; . ./$(CONSOLE_ENV); set +a; \
	site_id=$$(curl -sk -H "X-API-Key: $$UNIFI_API_KEY" https://localhost:11443/proxy/network/integrations/v1/sites \
		| python3 -c "import json,sys; print(json.load(sys.stdin)['data'][0]['id'])" 2>/dev/null); \
	if [ -z "$$site_id" ]; then echo "Could not resolve a site ID — is the console up? Try 'make console-up'."; exit 1; fi; \
	TF_ACC=1 \
	UNIFI_API_KEY="$$UNIFI_API_KEY" \
	UNIFI_API_URL="https://localhost:11443/proxy/network/integrations" \
	UNIFI_SITE_ID="$$site_id" \
	UNIFI_ALLOW_INSECURE=true \
	go test -count $(TEST_COUNT) -timeout $(TEST_TIMEOUT) -run TestAcc -v $(ACCTEST_PKG) $(TESTARGS)

.PHONY: console-up
console-up:
	cd $(CONSOLE_DIR) && docker compose up -d

.PHONY: console-down
console-down:
	cd $(CONSOLE_DIR) && docker compose down

.PHONY: console-logs
console-logs:
	cd $(CONSOLE_DIR) && docker compose logs -f

# adopt-fleet adopts every device the unifi-emu service (started by
# console-up) currently reports as pending, using nothing but the API key —
# see test-infra/unifi/adopt_fleet.py and README.md's "Adopted device fleet"
# section. Run once after console-up and the one-time manual setup in
# README.md; safe to re-run (a no-op once everything is already adopted, and
# an error if unifi-emu hasn't started informing yet).
.PHONY: adopt-fleet
adopt-fleet:
	@test -f $(CONSOLE_ENV) || { echo "Missing $(CONSOLE_ENV) — see $(CONSOLE_DIR)/README.md to create one."; exit 1; }
	@set -a; . ./$(CONSOLE_ENV); set +a; python3 $(CONSOLE_DIR)/adopt_fleet.py

# enable-zbf force-enables Zone Based Firewall directly in the console's own
# database — there's no API/UI path to do this without real UDM-class
# console hardware. See test-infra/unifi/README.md's "Zone Based Firewall"
# section for the investigation that led here and exactly what this writes.
# Idempotent: safe to run on every CI job, a no-op once already enabled.
.PHONY: enable-zbf
enable-zbf:
	@$(CONSOLE_DIR)/enable_zbf.sh

# ci-test is the single entry point a CI job runs against an already-running,
# already-onboarded console (see README.md — the one-time human setup that
# creates it is not automated here on purpose; see its "Zone Based Firewall"
# and "Adopted device fleet" sections for why). Every step is idempotent, so
# re-running this against the same long-lived console is always safe.
.PHONY: ci-test
ci-test: adopt-fleet enable-zbf testacc
