#!/bin/bash
FOLDER="$(realpath "$(dirname "$0")")"

function usage() {
	echo "docker-entrypoint.sh {cmd} {arg1} ... {argN}"
	echo "  {cmd}    				command you want run"
	echo "  {arg1} ... {argN}		Command Arguments arg1 ... argN"
}

if [ "-h" = "$1" ] || [ "--help" = "$1" ]; then
	echo "Usage:"
	echo -e "$(usage)"
	echo "Re-Bind: Exit!!"
	exit 0
fi

echo "Re-Bind: Starting Re-Bind Container ..."
chown -R rebind:rebind /var/rebind
chown -R rebind:rebind /etc/rebind
chmod -Rf 0660 /var/rebind
chown -Rf 0660 /etc/rebind
echo "Updating Re-Web and Re-Bind services ..."
go get -u github.com/hellgate75/rebind/rebind
go get -u github.com/hellgate75/rebind/reweb
echo "Starting Re-Web and Re-Bind services ..."
service rebind restart
service reweb restart
echo "Re-Web and Re-Bind services start complete!!"
if [ $# -gt 0 ]; then
	echo "Re-Bind: Running command: $@"
	sh -c "$@"
fi 
tail -f /var/log/rebind/rebind.log
echo "Re-Bind: Container Exit!!"
exit 0