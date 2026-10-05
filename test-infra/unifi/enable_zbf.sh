#!/bin/sh
# Force-enables Zone Based Firewall on every non-hidden site, directly in
# the console's own database — there is no API/UI path to do this on a
# UniFi OS Server with no real UDM-class console hardware. See
# README.md's "Zone Based Firewall" section for the full investigation
# that led here (decompiling /unifi/app/lib/ace.jar) and exactly why this
# is necessary and safe.
#
# Idempotent and safe to run on every CI job: it checks each site's
# migration/zone state before writing anything, and only restarts the
# Network application (to clear its 1-hour zone cache) when it actually
# changed something. Needs CONTAINER (the running UniFi OS Server
# container; defaults to unifi-os-server) and docker exec access to it.
set -eu

CONTAINER="${CONTAINER:-unifi-os-server}"

changed=$(docker exec "$CONTAINER" mongo --port 27117 ace --quiet --eval '
function ensureZoneBasedFirewall(siteId) {
  var already = db.site_feature_migration.findOne({site_id: siteId, feature: "ZONE_BASED_FIREWALL"});
  if (already) {
    return false;
  }

  // The real migration this stands in for creates one system-default zone
  // per zone_key before marking the site migrated — skipping that leaves
  // zone-dependent logic (e.g. any firewall policy create) throwing
  // "Could not find hotspot firewall zone", since it always looks up the
  // Hotspot zone for network-purpose bookkeeping regardless of which zone
  // is actually being touched. Shape/keys/names verified against
  // com.ubnt.service.firewallzone.rCTEZFVxHHljb in ace.jar.
  var zones = [
    {zone_key: "internal", name: "Internal"},
    {zone_key: "external", name: "External"},
    {zone_key: "gateway",  name: "Gateway"},
    {zone_key: "vpn",      name: "VPN"},
    {zone_key: "hotspot",  name: "Hotspot"},
    {zone_key: "dmz",      name: "DMZ"},
    {zone_key: "mgmt",     name: "Management"}
  ];
  zones.forEach(function(z) {
    if (db.firewall_zone.findOne({site_id: siteId, zone_key: z.zone_key})) {
      return;
    }
    db.firewall_zone.insertOne({
      external_id: UUID(),
      site_id: siteId,
      name: z.name,
      zone_key: z.zone_key,
      default_zone: true,
      attr_no_edit: true
    });
  });

  db.site_feature_migration.insertOne({
    site_id: siteId,
    feature: "ZONE_BASED_FIREWALL",
    timestamp: NumberLong(Date.now())
  });
  return true;
}

var anyChanged = false;
db.site.find({attr_hidden: {$ne: true}}).forEach(function(site) {
  if (ensureZoneBasedFirewall(site._id.valueOf())) {
    print("enabled Zone Based Firewall for site " + site.name + " (" + site._id.valueOf() + ")");
    anyChanged = true;
  } else {
    print("Zone Based Firewall already enabled for site " + site.name + " (" + site._id.valueOf() + ")");
  }
});
print(anyChanged ? "CHANGED" : "UNCHANGED");
' 2>&1)

echo "$changed"

if [ "$(echo "$changed" | tail -1)" = "CHANGED" ]; then
  # The zone list cache (com.ubnt.service.firewallzone.YfrfkymPjtCnsw) is an
  # in-JVM Caffeine cache with a 1-hour TTL, invalidated only by the service
  # layer we just bypassed — restart the app to pick up the new zones
  # immediately instead of waiting it out.
  echo "restarting the Network application to clear its zone cache..."
  docker exec "$CONTAINER" systemctl restart unifi.service

  echo "waiting for the API to come back..."
  for i in $(seq 1 30); do
    if docker exec "$CONTAINER" curl -sk -o /dev/null -w '%{http_code}' https://localhost/proxy/network/integrations/v1/sites 2>/dev/null | grep -qE '^(200|401)$'; then
      echo "API is back up"
      exit 0
    fi
    sleep 2
  done
  echo "timed out waiting for the API to come back after restart" >&2
  exit 1
fi
