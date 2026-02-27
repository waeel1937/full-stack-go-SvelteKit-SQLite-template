#!/bin/sh
set -e
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/edge-app ./cmd/edge-app
