#!/usr/local/bin/bash

$(aws ecr get-login --no-include-email)
docker-compose pull
HOSTNAME=${HOSTNAME} docker-compose up -d --build
docker system prune -af
