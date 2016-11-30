#!/bin/sh
sudo locale-gen zh_CN.UTF-8
sudo locale-gen zh_CN.UTF-8 en_US.UTF-8
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.7.3.linux-amd64.tar.gz
echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.profile
. ~/.profile