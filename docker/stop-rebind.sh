#!/usr/bin/env bash

if [ -e /run/rebind/rebind.pid ]; then
    pid="$(cat /run/rebind/rebind.pid)"
    if [ -n "$pid" ]; then
            sig=0
            n=1
            while kill -$sig $pid 2>/dev/null; do
            if [ $n -eq 1 ]; then
                echo "waiting for pid $pid to die"
            fi
            if [ $n -eq 11 ]; then
                echo "giving up on pid $pid with kill -0; trying -9"
                sig=9
            fi
            if [ $n -gt 20 ]; then
                echo "giving up on pid $pid"
                break
            fi
            n=$(($n+1))
            sleep 1
            done
    fi
    pid = "$(ps -eaf|grep rebind|head -1|awk 'BEGIN {FS=OFS=" "}{print $2}')"
    while [ "" != "$pid" ]; do
        kill -9 $pid
        pid = "$(ps -eaf|grep rebind|head -1|awk 'BEGIN {FS=OFS=" "}{print $2}')"
        sleep 5
    done
    rm -f /run/rebind/rebind.pid
else
    echo "PID file doesn't exist. Service is stopped!!"
fi
