#!/bin/sh
set -eu

cd "${AIR_COMPOSE_WORKING_DIR:-/data}"
exec "$@"
