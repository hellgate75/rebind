#!/usr/bin/env bash

function savePid() {
    sleep 10
    echo "$PID" > /run/rebind/reweb.pid
}

reweb $@ > /var/log/reweb.service.log &
PID="$!"
echo "Re-Web PID: $PID"
savePid "$PID" &