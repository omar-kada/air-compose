#!/usr/bin/env bash
# Puts the machine back to a clean state and pre-pulls images
# (so no download progress bars end up in the recording).
set -euo pipefail
cd "$(dirname "$0")"
set -a; source ./demo.env; set +a

# 1. stop AirCompose FIRST (so it can't redeploy things mid-cleanup),
#    then everything it deployed
if [ -f "$DEMO_BASE/services/air-compose/compose.yaml" ]; then
  docker compose -f "$DEMO_BASE/services/air-compose/compose.yaml" down -v || true
fi
for f in "$DEMO_BASE"/services/*/compose.yaml; do
  [ -f "$f" ] && docker compose -f "$f" down -v || true
done

# 2. recreate an empty base dir owned by the current user
#    (files written by the AirCompose container are root-owned -> sudo)
sudo rm -rf "$DEMO_BASE"
sudo mkdir -p "$DEMO_BASE"
sudo chown "$(id -un)" "$DEMO_BASE"

echo "reset done"
