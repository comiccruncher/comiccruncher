#!/usr/local/bin/bash

set -e

eval $(~/.local/bin/aws ecr get-login --no-include-email)
docker-compose pull
HOSTNAME=${HOSTNAME} docker-compose up -d --build --remove-orphans
docker system prune -af
docker logout
