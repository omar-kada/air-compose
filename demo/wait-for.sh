#!/usr/bin/env bash
# Waits until the AirCompose UI answers.
cd "$(dirname "$0")"
set -a; source ./demo.env; set +a
for i in $(seq 1 90); do
  curl -sf "$UI_URL" >/dev/null && { echo "UI is up"; exit 0; }
  sleep 1
done
echo "AirCompose UI did not come up at $UI_URL" >&2
exit 1
