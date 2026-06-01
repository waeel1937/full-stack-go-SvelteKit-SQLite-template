#!/bin/sh
set -e
mkdir -p dist
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -o dist/edge-app ./cmd/edge-app
