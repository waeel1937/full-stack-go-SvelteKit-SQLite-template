#!/bin/sh
set -e

docker compose -f edge-app/deploy/docker/docker-compose.yml up -d --build
sleep 6
docker inspect --format='{{.State.Health.Status}}' edge-app
docker compose -f edge-app/deploy/docker/docker-compose.yml down
