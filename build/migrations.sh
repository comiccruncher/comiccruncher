#!/usr/bin/env bash

exec 1> >(logger -s -t $(basename $0)) 2>&1

~/bin/migrations

