package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/dgamingfoundation/dwh/x/auctionHelper"

	common "github.com/dgamingfoundation/dwh/x/common"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	common.InitConfig()
	log.SetLevel(log.DebugLevel)
	var ctx = context.Background()

	if viper.GetBool(common.PprofEnabledFlag) {
		go func() {
			log.Println(http.ListenAndServe(viper.GetString(common.PprofHostPortFlag), nil))
		}()
	}
	hlprCfg := common.ReadCommonConfig(common.DefaultConfigName, common.DefaultConfigPath)

	db, err := common.GetDB(hlprCfg)
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()

	hlpr := auctionHelper.NewAuctionHelper(ctx, hlprCfg, db)

	hlpr.Run()
}
