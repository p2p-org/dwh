package main

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cliContext "github.com/dgamingfoundation/dkglib/lib/client/context"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/dgamingfoundation/dwh/indexer"
	app "github.com/dgamingfoundation/marketplace"
	mptypes "github.com/dgamingfoundation/marketplace/x/marketplace/types"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os/user"
	"path"
)

const (
	indexerStatePath = "./indexer.state"
)

const (
	nodeEndpoint  = "tcp://localhost:26657"
	chainID       = "mpchain"
	vfrHome       = ""
	height        = 0
	trustNode     = false
	broadcastMode = "sync"
	genOnly       = false
	validatorName = "user1"
)

var (
	cliHome = "~/.mpcli"
)

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	cliHome = path.Join(usr.HomeDir, "/", ".mpcli")
}

func main() {
	log.SetLevel(log.DebugLevel)
	var ctx = context.Background()

	db, err := common.GetDB()
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()
	db, err = common.InitDB(db, true)
	if err != nil {
		log.Fatalf("failed to InitDB: %v", err)
	}

	cliCtx, txDecoder, err := getEnv()
	if err != nil {
		log.Fatalf("failed to get env: %v", err)
	}
	idxrCfg := &indexer.Config{
		StatePath: indexerStatePath,
	}
	idxr, err := indexer.NewIndexer(ctx, idxrCfg, cliCtx, txDecoder, db,
		map[string]indexer.MsgHandler{
			mptypes.RouterKey: indexer.NewMarketplaceHandler(db, cliCtx),
		},
	)
	if err != nil {
		log.Fatalf("failed to create new indexer: %v", err)
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

func getEnv() (cliContext.CLIContext, sdk.TxDecoder, error) {
	cdc := app.MakeCodec()

	cliCtx, err := cliContext.NewCLIContext(chainID, nodeEndpoint, validatorName, genOnly, broadcastMode, vfrHome, height, trustNode, cliHome, "")
	if err != nil {
		return cliContext.CLIContext{}, nil, err
	}
	cliCtx = cliCtx.WithCodec(cdc)

	return cliCtx, auth.DefaultTxDecoder(cdc), nil
}
