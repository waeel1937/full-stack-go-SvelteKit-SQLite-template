#!/bin/sh
set -e

echo "running docker integration tests"
./edge-app/tests/integration.sh
