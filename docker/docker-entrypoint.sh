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
	echo "BIND9: Exit!!"
	exit 0
fi

echo "BIND9: Starting Bind9 Container ..."

if [ "" = "$ZONE_NAME" ]; then
	echo "BIND9: Zone cannot be empty calue ..."
	echo "BIND9: Exit!!"
	exit 0
fi

if [ ! -e /root/.installed ] || [ "ever" = "$REBUILD_CONFIG" ]; then
	if [ "yes" = "$USE_CONFIG_FROM_VOLUME" ]; then
		echo "BIND9: You decided to provide custom config files from the shared volume ..."
	else
		echo "BIND9: Mofifying template files ..."
		echo "ZONE: $MY_ZONE"
		echo "FORWAREDS: $FORWARDERS"
		if [ "" = "$FORWARDERS" ]; then
			FORWARDERS="8.8.8.8;8.8.4.4"
		fi
		if [[ $FORWARDERS =~ ';'$ ]]; then 
			FWDTEXT="${FORWARDERS}"
		else 
			FWDTEXT="${FORWARDERS};"
		fi
		FWDTEXT="$(echo $FWDTEXT|sed 's/;/;<CR>/g')"
		echo "BIND9: Replacing zone name in /etc/bind/named.conf.local ..."
		sed -i "s/MY_ZONE/$ZONE_NAME/g" /etc/bind/named.conf.local
		echo "BIND9: Replacing forwarders in /etc/bind/named.conf.options ..."
		sed -i "s/FWDTEXT/$FWDTEXT/g" /etc/bind/named.conf.options
		sed -i 's/<CR>/\n\t\t/g' /etc/bind/named.conf.options
		echo -e "$(cat /etc/bind/named.conf.options)" > /etc/bind/named.conf.options
		echo "BIND9: Spooling /etc/bind/named.conf.local:"
		cat /etc/bind/named.conf.local
		echo "BIND9: Spooling /etc/bind/named.conf.options:"
		cat /etc/bind/named.conf.options
		echo " "
		/etc/init.d/bind9 restart 2>&1 > /dev/null
	fi
	touch /root/.installed
else
	echo "BIND9: Initialization already performed, continue with container activation ..."
fi

if [ $# -gt 0 ]; then
	echo "BIND9: Running command: $@"
	sh -c "$@"
fi 
tail -f /var/log/bind/bind.log
echo "BIND9: Exit!!"
exit 0