#!/usr/bin/env bash

echo "Installing Re-Bind and Re-Web services ..."
go get -u github.com/hellgate75/rebind/rebind
go get -u github.com/hellgate75/rebind/reweb
chmod +x /root/go/bin/*
echo "Moving Re-Bind and Re-Web executables to /usr/sbin folder ..."
mv /root/go/bin/rebind /usr/sbin/
mv /root/go/bin/reweb /usr/sbin/

echo "Enable Re-Bind and Re-Web services ..."
update-rc.d rebind defaults
update-rc.d rebind enable 2
update-rc.d rebind enable 3
update-rc.d rebind enable 4
update-rc.d rebind enable 5
update-rc.d reweb defaults
update-rc.d reweb enable 2
update-rc.d reweb enable 3
update-rc.d reweb enable 4
update-rc.d reweb enable 5

/etc/init.d/rebind start
/etc/init.d/reweb start

rm -f /root/install-rebind.sh