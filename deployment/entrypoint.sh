#!/bin/sh
set -e

traefik --configFile=/etc/traefik/traefik.toml &

docker-compose -f /go/src/app/deployment/compose.yaml up -d &

/go/src/app/bin/sloth run -p 9090
