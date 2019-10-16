package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/dgamingfoundation/dwh/common"
	dwh_common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/dgamingfoundation/dwh/x/indexer"
	"github.com/dgamingfoundation/dwh/x/indexer/handlers"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
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

	db, err := common.GetDB()
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()

	cliCtx, txDecoder, err := handlers.GetEnv()
	if err != nil {
		log.Fatalf("failed to get env: %v", err)
	}

	idxrCfg := dwh_common.ReadCommonConfig(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)
	idxrCfg.StatePath = viper.GetString(common.StatePathFlag)
	idxr, err := indexer.NewIndexer(ctx, idxrCfg, cliCtx, txDecoder, db,
		indexer.WithHandler(handlers.NewMarketplaceHandler(cliCtx)),
	)
	if err != nil {
		log.Fatalf("failed to create new indexer: %v", err)
	}
	if err := idxr.Setup(idxrCfg.ResetDatabase); err != nil {
		log.Fatalf("failed to setup Indexer: %v", err)
	}

	if viper.GetBool(common.PrometheusEnabledFlag) {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(viper.GetString(common.PrometheusHostPortFlag), nil); err != nil {
				log.Fatalf("failed to run prometheus: %v", err)
			}
		}()
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		err := common.WaitInterrupted(ctx)
		idxr.Stop()
		return err
	})
	wg.Go(func() error {
		log.Info("starting indexer")
		defer log.Info("stopping indexer")
		return idxr.Start()
	})

	if err := wg.Wait(); err != nil {
		log.Fatalf("indexer stopped: %v", err)
	}
}
