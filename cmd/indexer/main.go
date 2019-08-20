package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os/user"
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cliContext "github.com/dgamingfoundation/dkglib/lib/client/context"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/dgamingfoundation/dwh/handlers"
	"github.com/dgamingfoundation/dwh/indexer"
	app "github.com/dgamingfoundation/marketplace"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

const (
	pprofEnabledFlag  = "pprof_enabled"
	pprofHostPortFlag = "pprof_host_port"
	statePathFlag     = "state_path"
	configFileName    = "indexer"
	nodeEndpointFlag  = "node_endpoint"
	chainIDFlag       = "chain_id"
	vfrHomeFlag       = "vfr_home"
	heightFlag        = "height"
	trustNodeFlag     = "trust_node"
	broadcastModeFlag = "broadcast_mode"
	genOnlyFlag       = "gen_only"
	userNameFlag      = "user_name"
	cliHomeFlag       = "cli_home"
)

func init() {
	initConfig()
}

func main() {
	log.SetLevel(log.DebugLevel)
	var ctx = context.Background()

	if viper.GetBool(pprofEnabledFlag) {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
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

	cliCtx, txDecoder, err := getEnv()
	if err != nil {
		log.Fatalf("failed to get env: %v", err)
	}
	idxrCfg := &indexer.Config{
		StatePath: viper.GetString(statePathFlag),
	}
	idxr, err := indexer.NewIndexer(ctx, idxrCfg, cliCtx, txDecoder, db,
		indexer.WithHandler(handlers.NewMarketplaceHandler(cliCtx)),
	)
	if err != nil {
		log.Fatalf("failed to create new indexer: %v", err)
	}
	if err := idxr.Setup(true); err != nil {
		log.Fatalf("failed to setup Indexer: %v", err)
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9080", nil); err != nil {
			log.Fatalf("failed to run prometheus: %v", err)
		}
	}()

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

func getEnv() (cliContext.CLIContext, sdk.TxDecoder, error) {
	cdc := app.MakeCodec()

	cliCtx, err := cliContext.NewCLIContext(
		viper.GetString(chainIDFlag),
		viper.GetString(nodeEndpointFlag),
		viper.GetString(userNameFlag),
		viper.GetBool(genOnlyFlag),
		viper.GetString(broadcastModeFlag),
		viper.GetString(vfrHomeFlag),
		viper.GetInt64(heightFlag),
		viper.GetBool(trustNodeFlag),
		viper.GetString(cliHomeFlag),
		"")
	if err != nil {
		return cliContext.CLIContext{}, nil, err
	}
	cliCtx = cliCtx.WithCodec(cdc)

	return cliCtx, auth.DefaultTxDecoder(cdc), nil
}

func initConfig() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user, exiting: %v", err)
	}

	viper.SetDefault(pprofEnabledFlag, true)
	viper.SetDefault(pprofHostPortFlag, "localhost:6061")
	viper.SetDefault(statePathFlag, "./indexer.state")
	viper.SetDefault(nodeEndpointFlag, "tcp://localhost:26657")
	viper.SetDefault(chainIDFlag, "mpchain")
	viper.SetDefault(vfrHomeFlag, "")
	viper.SetDefault(heightFlag, 0)
	viper.SetDefault(trustNodeFlag, false)
	viper.SetDefault(broadcastModeFlag, "sync")
	viper.SetDefault(genOnlyFlag, false)
	viper.SetDefault(userNameFlag, "user1")
	viper.SetDefault(cliHomeFlag, path.Join(usr.HomeDir, ".mpcli"))

	viper.SetConfigName(configFileName)
	viper.AddConfigPath("$HOME/.dwh/cfg")
	viper.AddConfigPath(".")

	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Println("config file not found, using default configuration")
	} else {
		log.Fatalf("failed to parse config file, exiting: %v", err)
	}
}
