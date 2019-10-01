package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/imgservice"
)

func main() {
	worker, err := imgservice.NewImageProcessingWorker("defcfg", "~/.")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run worker")
	log.Println(worker.Run())
}
