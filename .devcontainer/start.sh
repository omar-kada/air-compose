#!/bin/sh
# air-compose startup script for Codespaces devcontainer
# Ensures config file, directories, and app all start correctly

set -e

# Create required directories
mkdir -p /data /data/repo /data/db

# Ensure config file exists at /data/config.yaml
# (Mount may fail in Codespaces; copy as fallback)
if [ ! -f /data/config.yaml ]; then
  cp .devcontainer/config.yaml /data/config.yaml 2>/dev/null || true
fi

# Show config for debugging
echo "=== Config file ==="
ls -la /data/config.yaml
echo "--- content ---"
cat /data/config.yaml
echo "---"

# Show relevant env vars
echo "=== Environment ==="
env | grep AIR_COMPOSE
echo "---"

# Start air-compose
echo "=== Starting air-compose ==="
cd /data
nohup /app/air-compose run > /tmp/air-compose.log 2>&1 &
echo "PID: $!"

sleep 5

# Show log
echo "=== air-compose log ==="
cat /tmp/air-compose.log

# Check port
echo "=== Port check ==="
if curl -s http://localhost:5005 > /dev/null 2>&1; then
  echo "Port 5005 is open ✅"
else
  echo "Port 5005 is NOT open ❌"
fi
