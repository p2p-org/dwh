package main

import (
	"github.com/dgamingfoundation/dwh/common"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	db, err := common.GetDB()
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()

	db = common.CreateTables(db)
	nfts := common.PopulateMockNFTs(15)
	for _, nft := range nfts {
		db = db.Create(nft)
	}
}
