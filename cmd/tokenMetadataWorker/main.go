package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/tokenMetadataService"
)

func main() {
	worker, err := tokenMetadataService.NewTokenMetadataWorker("", "~/.")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run worker")
	log.Println(worker.Run())
}
