#!/bin/sh
# air-compose startup script for Codespaces devcontainer
# Checks if app is already running (from Docker CMD), otherwise starts it.

set -e

# Create required directories
mkdir -p /data /opt/stacks /data/db

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

# Docker daemon check
echo "=== Docker check ==="
if docker info > /dev/null 2>&1; then
  echo "Docker daemon is accessible ✅"
  echo "Docker server version: $(docker version --format '{{.Server.Version}}' 2>/dev/null || echo 'unknown')"
else
  echo "Docker daemon is NOT accessible ❌"
  echo "Check: /var/run/docker.sock mount and user permissions"
fi

# Docker Compose check (air-compose uses 'docker compose config --services' for health checks)
echo "=== Docker Compose check ==="
if docker compose version > /dev/null 2>&1; then
  echo "docker compose is available ✅"
  docker compose version
else
  echo "docker compose is NOT available ❌"
  echo "Health checks will fail — containers may show as unhealthy"
fi
echo "---"

# Check if air-compose is already running (e.g., from Docker CMD)
echo "=== Starting air-compose ==="
if curl -s http://localhost:5005 > /dev/null 2>&1; then
  echo "Port 5005 is already open ✅ — air-compose is running"
else
  echo "Port 5005 is closed — starting air-compose..."
  cd /data
  nohup /app/air-compose run > /tmp/air-compose.log 2>&1 &
  echo "PID: $!"
  sleep 5
fi

# Show app log
echo "=== air-compose log ==="
cat /tmp/air-compose.log 2>/dev/null || echo "(no log file — app started by Docker CMD)"

# Check Stirling PDF port too
echo "=== Port check ==="
if curl -s http://localhost:5005 > /dev/null 2>&1; then
  echo "Port 5005 (air-compose) is open ✅"
else
  echo "Port 5005 (air-compose) is NOT open ❌"
fi
if curl -s http://localhost:8080 > /dev/null 2>&1; then
  echo "Port 8080 (Stirling PDF) is open ✅"
else
  echo "Port 8080 (Stirling PDF) is NOT open — may not be deployed yet"
fi