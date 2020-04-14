#!/bin/sh
echo "Installing packages ..."
install-pkgs vim wget git build-essential openssh-server
echo "Creating the GoPath ..."
mkdir -p /root/go/bin
echo "Creating the ssh root folder ..."
mkdir -p /root/.ssh
echo "Creating ssh root keys ..."
echo -e "\n\n\n" > ssh-keygen -t rsa 
ssh-keyscan github.com >> githubKey
ssh-keygen -lf githubKey
cat githubKey > ~/.ssh/known_hosts
chmod 600 /root/.ssh/*

