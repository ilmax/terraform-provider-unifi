#!/bin/sh
# Re-creates the Zone Based Firewall state a fresh Network 11 console starts with.
#
# Why: adopting an emulated gateway makes Network 11 disable ZBF for the site
# ("Zone-Based Firewall is being disabled for Site[..] because it is not
# supported") because unifi-emu's inform payload has no
# udapi_version.versionDetail["firewall/filter"|"firewall/mangle"] (needs >= 4
# and >= 6). The controller removes the site's firewall_zone rows and its
# config.site_feature_migration ZONE_BASED_FIREWALL row; this puts them back
# (same shape as a fresh console) and restarts the Network application to drop
# its cached zone list.
#
# Network 11 keeps its data in Postgres (unifi-network database on port 5434),
# not MongoDB, so the old enable_zbf.sh (Mongo) does not apply.
#
# Idempotent: does nothing when the site already has its zones. Needs CONTAINER
# (default uos-test) and docker exec access.
set -eu

CONTAINER="${CONTAINER:-${UOS_CONTAINER:-uos-test}}"
psql_in() { docker exec -i "$CONTAINER" su postgres -c "psql -p 5434 -d unifi-network -At -v ON_ERROR_STOP=1"; }

count() { echo "SELECT count(*) FROM config.firewall_zone" | psql_in; }
before=$(count)

psql_in <<'SQL' >/dev/null
DO $$
DECLARE s uuid;
BEGIN
  FOR s IN SELECT id FROM config.site WHERE name <> 'super' LOOP
    IF EXISTS (SELECT 1 FROM config.firewall_zone WHERE site_id = s) THEN CONTINUE; END IF;
    INSERT INTO config.firewall_zone (site_id, external_id, name, zone_key, default_zone, no_edit) VALUES
      (s, gen_random_uuid(), 'Internal', 'internal', true, false),
      (s, gen_random_uuid(), 'External', 'external', true, true),
      (s, gen_random_uuid(), 'Gateway',  'gateway',  true, true),
      (s, gen_random_uuid(), 'Vpn',      'vpn',      true, true),
      (s, gen_random_uuid(), 'Hotspot',  'hotspot',  true, false),
      (s, gen_random_uuid(), 'Dmz',      'dmz',      true, false);
    INSERT INTO config.site_feature_migration (site_id, feature, origin, "timestamp")
      SELECT s, 'ZONE_BASED_FIREWALL', 'AUTOMATIC', (extract(epoch from now()) * 1000)::bigint
      WHERE NOT EXISTS (SELECT 1 FROM config.site_feature_migration WHERE site_id = s AND feature = 'ZONE_BASED_FIREWALL');
  END LOOP;
END $$;
SQL

if [ "$(count)" != "$before" ]; then
  echo "Restored Zone Based Firewall zones; restarting the Network application to clear its zone cache"
  docker exec "$CONTAINER" systemctl restart unifi.service
  # Wait for the Integrations API to answer again, then for every device to
  # be back ONLINE (state 1): right after a restart they sit in state 5 with no
  # port table, and tests that read gateway ports/WANs would hit that window.
  base="${UOS_URL:-https://localhost:11443}/proxy/network"
  i=0
  until [ "$(curl -ks -o /dev/null -w '%{http_code}' -H "X-API-Key: ${UNIFI_API_KEY:-}" "$base/integrations/v1/info")" = 200 ]; do
    i=$((i + 1)); [ "$i" -gt 60 ] && { echo "Network did not come back"; exit 1; }
    sleep 5
  done
  until curl -ks -H "X-API-Key: ${UNIFI_API_KEY:-}" "$base/api/s/default/stat/device-basic" \
    | python3 -I -c 'import json,sys; d=json.load(sys.stdin)["data"]; sys.exit(0 if d and all(x["state"]==1 for x in d) else 1)'; do
    i=$((i + 1)); [ "$i" -gt 120 ] && { echo "Devices did not come back online"; exit 1; }
    sleep 5
  done
else
  echo "Zone Based Firewall already in place"
fi
