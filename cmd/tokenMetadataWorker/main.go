package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/tokenMetadataService"
)

func main() {
	worker, err := tokenMetadataService.NewTokenMetadataWorker("config", "/root/")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run tokenMetaData worker")
	log.Println(worker.Run())
}
