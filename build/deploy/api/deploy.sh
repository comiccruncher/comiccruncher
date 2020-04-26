#!/usr/local/bin/bash

set -e

eval $(~/.local/bin/aws ecr get-login --no-include-email)

docker-compose pull

docker run --rm --env-file=.env 570480763436.dkr.ecr.us-east-1.amazonaws.com/comiccruncher/tasks:latest migrations

docker-compose up -d --build --remove-orphans

docker system prune -af

docker logout
