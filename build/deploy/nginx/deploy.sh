#!/bin/bash

HOSTNAME=${HOSTNAME} docker-compose up -d --build
docker system prune -af
