package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/x/imgresizer"
)

func main() {
	worker, err := imgresizer.NewImageProcessingWorker("config", "/root/")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run image worker")
	log.Println(worker.Run())
}
