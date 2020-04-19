#!/usr/bin/env bash
if [[ "" = "$GOVER" ]]; then
    export GOVER="$(wget -q -O - https://golang.org/doc/devel/release.html | grep "<h2 id=\"go" | awk 'BEGIN {FS=OFS=" "}{print $2}' | awk 'BEGIN {FS=OFS="\""}{print $2}'|head -1|awk 'BEGIN {FS=OFS="go"}{print $2}')"
    echo "Latest detected Go Language version is $GOVER"
else
    echo "Using provided Go Language version: $GOVER"
fi
export ARCH=amd64
export OS=linux
GIT_USER=hellgate75
GIT_EMAIL=hellgate75@gmail.com
rm -f /root/install-golang-prereq.sh
wget -Lq https://dl.google.com/go/go${GOVER}.${OS}-${ARCH}.tar.gz -O /root/go${GOVER}.${OS}-${ARCH}.tar.gz
tar -xzf /root/go${GOVER}.${OS}-${ARCH}.tar.gz -C /usr/share/
rm -Rf /root/go${GOVER}.${OS}-${ARCH}.tar.gz
chmod 777 ${GOINST}/bin/*
git config --global user.name "${GIT_USER}"
git config --global user.email "${GIT_EMAIL}"
go get -u  github.com/mdempsky/gocode &&\
go get -u golang.org/x/tools/... &&\
go get -u github.com/golang/dep/cmd/dep &&\
go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
ln -s -T /root/go/bin/golangci-lint /root/go/bin/golint &&\
go version
rm -f /root/install-golang.sh
