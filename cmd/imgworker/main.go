package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/imgservice"
)

func main() {
	worker, err := imgservice.NewImageProcessingWorker("config", "/root/")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run image worker")
	log.Println(worker.Run())
}
