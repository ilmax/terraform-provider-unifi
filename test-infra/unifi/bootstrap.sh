#!/bin/sh
# Prepares a freshly started console for the acceptance tests and writes
# test-infra/unifi/.env:
#
#   1. creates the local admin account (POST /api/setup, no cloud account),
#   2. installs the UniFi Network application (install-network.sh),
#   3. creates an API key (the call the web UI makes on Integrations > Create
#      New API Key) and saves it as UNIFI_API_KEY.
#
# Idempotent: when .env already holds a key that the console accepts it does
# nothing. A console that was set up by hand, or whose .env was lost, cannot be
# recovered from here; reset it with `make console-reset`.
#
#   UOS_URL        console URL (default https://localhost:11443)
#   UOS_CONTAINER  console container (default uos-test)
#   NETWORK_VERSION / NETWORK_URL / NETWORK_CHANNEL   see install-network.sh
set -eu

here=$(cd "$(dirname "$0")" && pwd)
env_file=$here/.env
: "${UOS_URL:=https://localhost:11443}"
: "${UOS_CONTAINER:=uos-test}"
export UOS_CONTAINER

cookies=$(mktemp)
trap 'rm -f "$cookies"' EXIT

# json <python expression over d> reads JSON from stdin.
json() { python3 -I -c 'import json,sys; d=json.load(sys.stdin); print(eval(sys.argv[1]))' "$1"; }

env_get() { [ -f "$env_file" ] && sed -n "s/^$1=//p" "$env_file" | tail -1 || true; }

env_set() {
  touch "$env_file"
  grep -v "^$1=" "$env_file" > "$env_file.tmp" || true
  printf '%s=%s\n' "$1" "$2" >> "$env_file.tmp"
  mv "$env_file.tmp" "$env_file"
  chmod 600 "$env_file"
}

# The Integrations API answers on /proxy/network/integrations on the consoles
# this repo used so far; probe the usual spellings instead of assuming one.
integrations_url() {
  for p in integrations integration; do
    if [ "$(curl -ks -o /dev/null -w '%{http_code}' -H "X-API-Key: $1" "$UOS_URL/proxy/network/$p/v1/info")" = 200 ]; then
      echo "$UOS_URL/proxy/network/$p"
      return 0
    fi
  done
  return 1
}

echo "Waiting for UniFi OS on $UOS_URL ..."
i=0
until curl -ks "$UOS_URL/api/system" 2>/dev/null | python3 -I -c 'import json,sys; json.load(sys.stdin)' 2>/dev/null; do
  i=$((i + 1)); [ $i -le 90 ] || { echo "UniFi OS did not answer in time" >&2; exit 1; }
  sleep 5
done

key=$(env_get UNIFI_API_KEY)
if [ -n "$key" ] && integrations_url "$key" >/dev/null; then
  echo "Console already set up: the API key in $env_file works."
  exit 0
fi

is_setup=$(curl -ks "$UOS_URL/api/system" | json 'd.get("isSetup")')
password=$(env_get UOS_CONSOLE_PASSWORD)

if [ "$is_setup" != True ]; then
  # UniFi OS wants mixed case, a digit and a symbol; the prefix guarantees them.
  password="Aa1-$(openssl rand -hex 16)"
  status=$(curl -ks -o /dev/null -w '%{http_code}' -X POST "$UOS_URL/api/setup" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"acctest\",\"timezone\":\"Etc/UTC\",\"optimizeNetwork\":false,\"sendDiagnostics\":\"anonymous\",\"newAccount\":false,\"username\":\"admin\",\"password\":\"$password\"}")
  [ "$status" = 200 ] || { echo "POST /api/setup failed: HTTP $status" >&2; exit 1; }
  env_set UOS_CONSOLE_PASSWORD "$password"
  echo "Created the admin account."
elif [ -z "$password" ]; then
  echo "The console is already set up but $env_file has no UOS_CONSOLE_PASSWORD or working UNIFI_API_KEY." >&2
  echo "Run \`make console-reset\` to start from a clean console." >&2
  exit 1
fi

"$here/install-network.sh"

# One login, then everything else uses the API key (UniFi OS rate-limits login).
status=$(curl -ks -c "$cookies" -o /dev/null -w '%{http_code}' -X POST "$UOS_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"$password\",\"remember\":false}")
[ "$status" = 200 ] || { echo "login failed: HTTP $status (the password in $env_file may be stale: make console-reset)" >&2; exit 1; }

csrf=$(curl -ks -b "$cookies" -D - -o /dev/null "$UOS_URL/api/users/self" | tr -d '\r' | awk -F': ' 'tolower($1)=="x-csrf-token"{print $2}')
user_id=$(curl -ks -b "$cookies" "$UOS_URL/api/users/self" | json 'd["unique_id"]')
key=$(curl -ks -b "$cookies" -H "X-CSRF-Token: $csrf" -H 'Content-Type: application/json' \
  -X POST "$UOS_URL/proxy/users/api/v2/user/$user_id/keys" -d '{"name":"acctest"}' | json 'd["data"]["full_api_key"]')
[ -n "$key" ] || { echo "no API key in the create response" >&2; exit 1; }
env_set UNIFI_API_KEY "$key"
echo "Created an API key."

i=0
until base=$(integrations_url "$key"); do
  i=$((i + 1)); [ $i -le 36 ] || { echo "the Integrations API did not answer" >&2; exit 1; }
  sleep 5
done
echo "Ready: Integrations API at $base ($(curl -ks -H "X-API-Key: $key" "$base/v1/info"))"
