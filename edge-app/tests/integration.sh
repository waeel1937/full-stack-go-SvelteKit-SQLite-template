#!/bin/sh
set -e

echo "starting docker stack"
docker compose -f edge-app/deploy/docker/docker-compose.yml up -d --build

echo "waiting for service"
sleep 6

echo "checking status endpoint"
curl -fs http://localhost:8080/api/v1/status >/dev/null

echo "checking aggregates endpoint"
curl -fs "http://localhost:8080/api/v1/aggregates?window_ms=1000" >/dev/null

echo "checking raw endpoint"
curl -fs http://localhost:8080/api/v1/raw >/dev/null

echo "checking grpc port"
nc -z localhost 9090

echo "all tests passed"

docker compose -f edge-app/deploy/docker/docker-compose.yml down
