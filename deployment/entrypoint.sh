#!/bin/sh
set -e

traefik --configFile=/etc/traefik/traefik.toml &

/go/src/app/bin/sloth run -p 9090
