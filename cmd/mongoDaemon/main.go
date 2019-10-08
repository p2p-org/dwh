package main

import (
	"log"

	"github.com/dgamingfoundation/dwh/x/mongoDaemon"
)

func main() {
	worker, err := mongoDaemon.NewMongoDaemon("config", "/root/")
	if err != nil {
		panic(err)
	}
	defer worker.Closer()

	log.Println("run mongo daemon")
	log.Println(worker.Run())
}
