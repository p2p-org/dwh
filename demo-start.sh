#!/bin/bash

if [ ! -d ~/data/dwh0 ]; then
  mkdir -p ~/data/dwh0
  cp ./config-ibc0.toml ~/data/dwh0/config.toml
fi
if [ ! -d ~/data/dwh1 ]; then
  mkdir -p ~/data/dwh1
  cp ./config-ibc1.toml ~/data/dwh1/config.toml
fi

docker build -t dwh_indexer --build-arg APPNAME=indexer .
docker network create dwh-network
docker-compose -f demo.yml up -d
