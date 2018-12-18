#!/usr/bin/env bash

# exec 1> >(logger -s -t $(basename $0)) 2>&1

ps -ef | awk '/webapp/ {print $2}' | xargs kill -9 > /dev/null

mv ~/bin/webapp1 ~/bin/webapp

nohup ~/bin/webapp start -p 8001 | logger &
