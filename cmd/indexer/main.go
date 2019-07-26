package main

import (
	"context"
	"os"
	"path"

	"github.com/dgamingfoundation/dwh/handlers"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/dgamingfoundation/dwh/indexer"
	app "github.com/dgamingfoundation/marketplace"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	"golang.org/x/sync/errgroup"
)

const (
	indexerStatePath = "./indexer.state"
)

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

	cliCtx, txDecoder := getEnv()
	idxrCfg := &indexer.Config{
		StatePath: indexerStatePath,
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

// TODO: use the simpler context that is developed by @pr0n00gler.
func getEnv() (cliContext.CLIContext, sdk.TxDecoder) {
	viper.Set("home", "~/.mpcli")
	viper.Set("chain-id", "mpchain")

	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "mpcli",
		Short: "marketplace Client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
	)

	return cliContext.NewCLIContext().WithCodec(cdc), auth.DefaultTxDecoder(cdc)
}

func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		client.LineBreak,
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByTagsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		client.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		client.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		client.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}

func populateTestData(db *gorm.DB) {
	nfts := common.PopulateMockNFTs(15)
	for _, nft := range nfts {
		db = db.Create(nft)
	}
}
