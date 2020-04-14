FROM ubuntu:bionic

ENV GROUP_NAME="aangine"\
 FORWARDERS="8.8.8.8;8.8.4.4"\
 DEBIAN_FRONTEND=noninteractive\
 USE_CONFIG_FROM_VOLUME=no\
 REBUILD_CONFIG=once

RUN apt-get update \
  && apt-get install -y apt-utils \
  tar \
  lzma \
  gzip \
  && apt-get install -y \
  curl \
  vim \
  dos2unix \
  net-tools \
  iproute2 &&\
 apt-get autoremove &&\
 rm -rf /var/lib/apt/lists/* &&\
 mkdir -p /var/rebind &&\
 mkdir -p /etc/rebind

# Copy configuration files
COPY config/rebind.yaml /etc/rebind
COPY config/reweb.yaml /etc/rebind
COPY logrotate/rebind /etc/logrotate.d
COPY logrotate/reweb /etc/logrotate.d
COPY defaults/rebind /etc/defaults
COPY defaults/web /etc/defaults
COPY init.d/rebind /etc/init.d
COPY init.d/reweb /etc/init.d

# Copy entry point
COPY install-golang.sh /root
COPY docker-entrypoint.sh /

# Enable Re-Bin and Re-Web
RUN dos2unix /root/install-golang.sh &&\
 chmod +x /root/install-golang.sh &&
 /root/install-golang.sh &&
 go get -u github.com/rebind/rebind &&\
 go get -u github.com/rebind/reweb &&\
 dos2unix /docker-entrypoint.sh &&\
 dos2unix /etc/logrotate.d/* &&\
 dos2unix /etc/rebind/* &&\
 dos2unix /etc/init.d/* &&\
 mkdir /var/log/bind &&\
 chown bind:bind /var/log/bind &&\
 update-rc.d rebind defaults &&\
 update-rc.d rebind enable &&\
 update-rc.d reweb defaults &&\
 update-rc.d reweb enable &&\
 chmod +x /docker-entrypoint.sh

VOLUME ["/etc/rebind", "/var/rebind"]

EXPOSE 9000/tcp 53/udp

#Entry point

ENTRYPOINT ["/docker-entrypoint.sh"]