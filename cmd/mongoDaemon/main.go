package main

import (
	"log"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"

	"github.com/dgamingfoundation/dwh/x/mongoDaemon"
)

func main() {
	worker, err := mongoDaemon.NewMongoDaemon(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run mongo daemon")
	log.Println(worker.Run())
}
