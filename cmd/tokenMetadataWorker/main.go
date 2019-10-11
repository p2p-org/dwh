package main

import (
	stdLog "log"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/dgamingfoundation/dwh/x/tokenMetadataService"
)

func main() {
	worker, err := tokenMetadataService.NewTokenMetadataWorker(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	stdLog.Println("run tokenMetaData worker")
	stdLog.Println(worker.Run())
}
