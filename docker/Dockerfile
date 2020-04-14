FROM ubuntu:bionic

ENV GOINST=/usr/share/go\
	GOPATH=/root/go
ENV PATH=${PATH}:${GOPATH}/bin:${GOINST}/bin\
 DEBIAN_FRONTEND=noninteractive

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
COPY defaults/rebind /etc/defaults
COPY defaults/web /etc/defaults
COPY init.d/rebind /etc/init.d
COPY init.d/reweb /etc/init.d

# Copy entry point
COPY install-golang.sh /root
COPY docker-entrypoint.sh /

# Enable Re-Bin and Re-Web
RUN dos2unix /root/install-golang.sh &&\
 chmod +x /root/install-golang.sh &&\
 sh -c /root/install-golang.sh &&\
 go get -u github.com/rebind/rebind &&\
 go get -u github.com/rebind/reweb &&\
 chmod +x /root/go/bin/* &&\
 mv /root/go/bin/reweb /usr/sbin/ &&\
 mv /root/go/bin/reweb /usr/sbin/ &&\
 dos2unix /docker-entrypoint.sh &&\
 dos2unix /etc/rebind/* &&\
 dos2unix /etc/init.d/* &&\
 mkdir /var/log/bind &&\
 chown bind:bind /var/log/bind &&\
 update-rc.d rebind defaults &&\
 update-rc.d rebind enable &&\
 update-rc.d reweb defaults &&\
 update-rc.d reweb enable &&\
 chmod +x /docker-entrypoint.sh

# Exposed volums : /etc/rebind: Configuration volume, /var/rebind: Data volume

VOLUME ["/etc/rebind", "/var/rebind"]

# Exposed ports : REST Service: 9000, Dns Service: 53

EXPOSE 9000/tcp 53/udp

# Entry point

ENTRYPOINT ["/docker-entrypoint.sh"]