#!/usr/bin/env bash


cur_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )

cd $cur_path

docker_img_worker_name=dwh_img_worker
docker_metadata_worker_name=dwh_tmd_worker
docker_img_storage_name=dwh_img_storage
docker_mongo_daemon_name=dwh_mongo_daemon
docker_indexer_name=dwh_indexer

if [ $# -ne 1 ]; then
    echo "Illegal number of parameters: $#"
    exit 0
fi

while test $# -gt 0; do
  case "$1" in
    help)
      echo "testnet - run randapp testnet"
      echo " "
      echo "testnet [options]"
      echo " "
      echo "options:"
      echo "help                        show brief help"
      echo "up                          create & start all containers; images must be built"
      echo "down                        removes all containers"
      echo "reset                       removes containers, create and starts them without rebuild; equals down && up"
      echo "start                       start all stopped containers without recreation"
      echo "stop                        stop all running containers without data loss"
      echo "rebuild                     rebuild dwh images:
                            imgstorage, imgworker, indexer, mongoDaemon, tokenMetadataWorker"
      echo "rebuild-mp                  rebuild marketplace image"
      echo "rebuild-all                 rebuild all docker images, including marketplace
                            IMPORTANT: marketplace src MUST be in ./../marketplace"
      echo "purge                       remove all containers, delete local files"
      exit 0
      ;;
    up)
      docker network create dwh-network
      docker-compose -f infrastructure-compose.yml up -d
      sleep 24
      #docker-compose -f dwh-compose.yml up -d --scale token_meta_data=2 --scale img_worker=2
      docker-compose -f dwh-compose.yml up -d
      exit 0
      ;;
    down)
      docker-compose -f dwh-compose.yml down -v
      docker-compose -f infrastructure-compose.yml down -v
      sleep 5
      docker volume prune -f
      exit 0
      ;;
    reset)
      $cur_path/$0 down
      $cur_path/$0 up
      exit 0
      ;;
    start)
      docker-compose -f infrastructure-compose.yml start
      docker-compose -f dwh-compose.yml start
#      $cur_path/$0 up
      exit 0
      ;;
    stop)
      docker-compose -f dwh-compose.yml stop
      docker-compose -f infrastructure-compose.yml stop
#      $cur_path/$0 down
      exit 0
      ;;
    restart)
      $cur_path/$0 stop
      $cur_path/$0 start
      exit 0
      ;;
    rebuild)
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

      docker build -t $docker_indexer_name --build-arg APPNAME=indexer .
      docker build -t $docker_img_storage_name --build-arg APPNAME=imgstorage .
      docker build -t $docker_metadata_worker_name --build-arg APPNAME=tokenMetadataWorker .
      docker build -t $docker_img_worker_name --build-arg	APPNAME=imgworker .
      docker build -t $docker_mongo_daemon_name --build-arg APPNAME=mongoDaemon .

      rm -rf $cur_path/vendor
      exit 0
      ;;
    rebuild-mp)
      $cur_path/$0 stop
      docker rm dwh_marketplace
      $cur_path/../marketplace/buildDocker.sh
      exit 0
      ;;
    rebuild-all)
      $cur_path/$0 rebuild-mp
      $cur_path/$0 rebuild

      exit 0
      ;;
    seed)
      docker cp gen_marketplace_data.sh dwh_marketplace:/go/src/github.com/corestario/marketplace
      docker exec -it dwh_marketplace bash /go/src/github.com/corestario/marketplace/gen_marketplace_data.sh

      exit 0
      ;;
    logs-i)
      docker-compose -f dwh-compose.yml logs -f --tail="500" indexer

      exit 0
      ;;
    logs-m)
      docker-compose -f infrastructure-compose.yml logs -f --tail="500" marketplace

      exit 0
      ;;
    purge)
      $cur_path/$0 down
      rm -fr $cur_path/vendor
      rm -fr $cur_path/volumes

      exit 0
      ;;
    cli)
      docker exec -it dwh_marketplace zsh
      exit 0
      ;;
    *)
      echo "wrong argument:"
      echo "$1"
      exit 0
      ;;
  esac
done
