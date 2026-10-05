#!/usr/bin/env bash
# Cron-friendly wrapper: run the API watch, log to file, alert on change.
# Paths are resolved relative to this script's own location, so it works from
# anywhere in the repo. Recommended crontab (every 6h) — use the absolute path of
# this file on your machine:
#   0 */6 * * * /path/to/tools/unifi-network-api-compare/watch/unifi-watch.sh >> /path/to/tools/unifi-network-api-compare/watch/watch.log 2>&1
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG="$DIR/watch.log"
cd "$DIR"

echo "===== $(date -u '+%Y-%m-%d %H:%M:%S UTC') check =====" | tee -a "$LOG"
if python3 check_unifi_api_versions.py; then
  echo "ok: no new API surface" | tee -a "$LOG"
else
  rc=$?
  if [ "$rc" = "1" ]; then
    echo "!!! CHANGE DETECTED — see output above. Run the comparison tool." | tee -a "$LOG"
  else
    echo "error (rc=$rc) running check" | tee -a "$LOG"
  fi
fi
echo
