#!/bin/bash

set -e

eval $(~/.local/bin/aws ecr get-login --no-include-email)

docker pull 570480763436.dkr.ecr.us-east-1.amazonaws.com/comiccruncher/tasks:latest

docker run --rm --env-file=.env comiccruncher/tasks:latest migrations

docker-compose up -d api

docker exec $(docker ps | awk '/nginx:1.15-alpine/{print $1}') nginx -s reload

docker system prune -af
