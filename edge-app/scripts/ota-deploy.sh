#!/bin/sh
set -e

DEVICE="$1"
if [ -z "$DEVICE" ]; then
	exit 1
fi

scp dist/edge-app "$DEVICE:/opt/edge-app/edge-app.new"
ssh "$DEVICE" "
mv /opt/edge-app/edge-app /opt/edge-app/edge-app.old || true
mv /opt/edge-app/edge-app.new /opt/edge-app/edge-app
systemctl restart edge-app
"
