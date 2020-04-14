#!/usr/bin/env bash

echo "Installing Re-Bind and Re-Web services ..."
go get -u github.com/rebind/rebind
go get -u github.com/rebind/reweb
chmod +x /root/go/bin/*
echo "Moving Re-Bind and Re-Web executables to /usr/sbin folder ..."
mv /root/go/bin/reweb /usr/sbin/
mv /root/go/bin/reweb /usr/sbin/

echo "Enable Re-Bind and Re-Web services ..."
update-rc.d rebind defaults
update-rc.d rebind enable
update-rc.d reweb defaults
update-rc.d reweb enable

rm -f /root/install-rebind.sh