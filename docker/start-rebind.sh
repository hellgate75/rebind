#!/usr/bin/env bash

function savePid() {
    sleep 10
    echo "$PID" > /run/rebind/rebind.pid
}
rebind $@ > /var/log/rebind.service.log &
PID="$!"
#echo "Re-Bind PID: $PID"
savePid "$PID" &