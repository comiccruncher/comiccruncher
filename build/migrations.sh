#!/usr/bin/env bash

exec 1> >(logger -s -t $(basename $0)) 2>&1

/usr/local/bin/migrations
