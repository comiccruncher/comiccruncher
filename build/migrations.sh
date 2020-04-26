#!/usr/bin/env bash

docker pull 570480763436.dkr.ecr.us-east-1.amazonaws.com/comiccruncher/tasks:latest

exec 1> >(logger -s -t $(basename $0)) 2>&1

docker run --env-file=.env comiccruncher/tasks:latest migrations
