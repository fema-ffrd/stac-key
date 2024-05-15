#!/bin/bash

set -euo pipefail

source .env

docker compose -f docker-compose.traefik.yml up -d 
# docker compose -f docker-compose.traefik.yml logs -f

docker compose -f docker-compose.keyval.yml up -d 
# docker compose -f docker-compose.keyval.yml logs -f

docker compose -f docker-compose.stac.yml up -d 
docker compose -f docker-compose.stac.yml logs -f