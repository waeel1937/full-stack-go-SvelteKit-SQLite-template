#!/bin/sh
set -e

cd edge-app/deploy/docker

docker compose down -v || true
docker compose build
docker compose up -d

sleep 5

curl -s http://localhost:8080/api/v1/status
echo
curl -s http://localhost:8080/api/v1/aggregates?window_ms=1000
echo
curl -s http://localhost:8080/api/v1/raw
echo
