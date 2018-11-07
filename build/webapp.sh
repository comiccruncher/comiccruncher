#!/usr/bin/env bash

exec 1> >(logger -s -t $(basename $0)) 2>&1

kill $(ps aux | grep webapp | awk '{print $2}')

mv ~/bin/webapp1 ~/bin/webapp

~/bin/webapp start -p 8001
