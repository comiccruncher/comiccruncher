#!/usr/local/bin/bash

set -e

docker exec $(docker ps | awk '/nginx:1.15-alpine/{print $1}') nginx -s reload

docker system prune -af

docker logout
