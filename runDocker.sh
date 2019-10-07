#!/usr/bin/env bash

docker_image_worker_name=dwh_img_worker
docker_metadata_worker_name=dwh_tmd_worker

cur_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )

cd $cur_path
rm -rf $cur_path/vendor

gopath=$(whereis go | grep -oP '(?<=go: )(\S*)(?= .*)' -m 1)
PATH=$gopath:$gopath/bin:$PATH

echo $GOBIN
echo "--> Ensure dependencies have not been modified"

GO111MODULE=on go mod verify
GO111MODULE=on go mod vendor
GO111MODULE=off

chmod 0777 ./go.sum
chmod -R 0777 ./vendor

docker build -t $docker_image_worker_name -f ./worker.Dockerfile .
docker build -t $docker_metadata_worker_name -f ./metaData.Dockerfile .


rm -rf $cur_path/vendor

docker-compose up --scale token_meta_data=2 --scale img_worker=2
