#!/bin/sh
# air-compose startup script for Codespaces devcontainer
# Checks if app is already running (from Docker CMD), otherwise starts it.

set -e

# Create required directories
mkdir -p /data /data/repo /data/db

# Ensure config file exists (fallback for failed mount)
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

# Check if air-compose is already running (e.g., from Docker CMD)
echo "=== Starting air-compose ==="
if curl -s http://localhost:5005 > /dev/null 2>&1; then
  echo "Port 5005 is already open - air-compose is running (from CMD)"
else
  echo "Port 5005 is closed - starting air-compose now..."
  cd /data
  nohup /app/air-compose run > /tmp/air-compose.log 2>&1 &
  echo "PID: $!"
  sleep 5
fi

# Show app log
echo "=== air-compose log ==="
cat /tmp/air-compose.log 2>/dev/null || echo "(no log file - app started by Docker CMD)"

# Final port check
echo "=== Port check ==="
if curl -s http://localhost:5005 > /dev/null 2>&1; then
  echo "Port 5005 is open"
else
  echo "Port 5005 is NOT open"
fi