package main

import (
	stdLog "log"

	dwh_common "github.com/corestario/dwh/x/common"
	"github.com/corestario/dwh/x/mongoDaemon"
)

func main() {
	worker, err := mongoDaemon.NewMongoDaemon(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	stdLog.Println("run mongo daemon")
	stdLog.Println(worker.Run())
}
