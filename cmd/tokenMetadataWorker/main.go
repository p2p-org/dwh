package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/tokenMetadataSaverService"
)

func main() {
	worker, err := tokenMetadataSaverService.NewTokenMetadataWorker("defcfg", "~/.")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run worker")
	log.Println(worker.Run())
}
