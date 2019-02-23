#!/bin/bash

docker-compose pull
HOSTNAME=${HOSTNAME} docker-compose up -d --build
docker system prune -af
