FROM ubuntu:bionic

ENV GOINST=/usr/share/go\
	GOPATH=/root/go
ENV PATH=${PATH}:/root/go/bin:/usr/share/go/bin\
 DEBIAN_FRONTEND=noninteractive

RUN  mkdir -p /var/rebind &&\
 mkdir -p /etc/rebind &&\
 mkdir -p /var/log/rebind

# Copy configuration files
COPY config/passwd /etc
COPY config/group /etc
COPY config/rebind.yaml /etc/rebind
COPY config/reweb.yaml /etc/rebind
COPY defaults/rebind /etc/default
COPY defaults/reweb /etc/default
COPY init.d/rebind /etc/init.d
COPY init.d/reweb /etc/init.d

# Copy entry point
COPY install-golang-prereq.sh /root
COPY install-golang.sh /root
COPY install-rebind.sh /root
COPY install-pkgs /bin
COPY docker-entrypoint.sh /
COPY start-rebind.sh /usr/sbin
COPY stop-rebind.sh /usr/sbin
COPY start-reweb.sh /usr/sbin
COPY stop-reweb.sh /usr/sbin


# Install dos2unix file converter
RUN apt-get update &&\
 apt-get install -y dos2unix &&\
 apt-get autoremove &&\
 rm -rf /var/lib/apt/lists/*

# Config user and install tools
RUN chmod 644 /etc/passwd /etc/group &&\
 dos2unix /bin/install-pkgs &&\
 dos2unix /usr/sbin/start-* &&\
 dos2unix /usr/sbin/stop-* &&\
 dos2unix /etc/default/* &&\
 chmod +x /etc/default/* &&\
 chmod +x /bin/install-pkgs &&\
 chmod +x /usr/sbin/start-* &&\
 chmod +x /usr/sbin/stop-* &&\
 dos2unix /etc/passwd &&\
 dos2unix /etc/group &&\
 mkdir -p /home/rebind &&\
 chmod 660 /home/rebind

# Install system packages
RUN install-pkgs apt-utils \
  tar \
  lzma \
  gzip \
  && install-pkgs \
  curl \
  vim \
  net-tools \
  iproute2

# Prepare files and install go-lang
RUN dos2unix /root/install-golang-prereq.sh &&\
 chmod +x /root/install-golang-prereq.sh &&\
 sh -c /root/install-golang-prereq.sh &&\
 dos2unix /root/install-golang.sh &&\
 chmod +x /root/install-golang.sh &&\
 sh -c /root/install-golang.sh &&\
 dos2unix /docker-entrypoint.sh &&\
 dos2unix /etc/rebind/* &&\
 chown -R rebind:rebind /home/rebind &&\
 chown -R rebind:rebind /etc/rebind &&\
 dos2unix /etc/init.d/* &&\
 chmod -Rf 0660 /var/rebind &&\
 chown -Rf 0660 /etc/rebind &&\
 chmod 777 /etc/init.d/* &&\
 chown -R rebind:rebind /var/rebind &&\
 chown rebind:rebind /var/log/rebind &&\
 chmod +x /docker-entrypoint.sh

# Install Re-Bind and Re-Web via go get and enable services
RUN  dos2unix /root/install-rebind.sh &&\
 chmod +x /root/install-rebind.sh &&\
 sh -c /root/install-rebind.sh

# Exposed volums : /etc/rebind: Configuration volume, /var/rebind: Data volume,
# /var/log/rebind: logs path in case you want increase the number of logs in the
# rotate process

VOLUME ["/etc/rebind", "/var/rebind", "/var/log/rebind"]

# Exposed ports : REST Service: 9000, Dns Service: 53

EXPOSE 9000/tcp 53/udp

# Entry point

ENTRYPOINT ["/docker-entrypoint.sh"]