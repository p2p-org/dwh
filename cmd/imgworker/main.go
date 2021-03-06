package main

import (
	stdLog "log"

	dwh_common "github.com/corestario/dwh/x/common"
	"github.com/corestario/dwh/x/imgresizer"
)

func main() {
	worker, err := imgresizer.NewImageProcessingWorker(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	stdLog.Println("run image worker")
	stdLog.Println(worker.Run())
}
