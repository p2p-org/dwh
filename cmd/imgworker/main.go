package main

import (
	"log"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"

	"github.com/dgamingfoundation/dwh/x/imgresizer"
)

func main() {
	worker, err := imgresizer.NewImageProcessingWorker(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run image worker")
	log.Println(worker.Run())
}
