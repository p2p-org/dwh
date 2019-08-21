package handlers

import (
	"fmt"

	sdk "github.com/dgamingfoundation/cosmos-sdk/types"
	"github.com/dgamingfoundation/cosmos-sdk/x/auth"
	cliContext "github.com/dgamingfoundation/dkglib/lib/client/context"
	"github.com/dgamingfoundation/dwh/common"
	app "github.com/dgamingfoundation/marketplace"
	"github.com/spf13/viper"
)

func GetEnv() (cliContext.CLIContext, sdk.TxDecoder, error) {
	cdc := app.MakeCodec()

	fmt.Println(
		viper.GetString(common.ChainIDFlag),
		viper.GetString(common.NodeEndpointFlag),
		viper.GetString(common.UserNameFlag),
		viper.GetBool(common.GenOnlyFlag),
		viper.GetString(common.BroadcastModeFlag),
		viper.GetString(common.VfrHomeFlag),
		viper.GetInt64(common.HeightFlag),
		viper.GetBool(common.TrustNodeFlag),
		viper.GetString(common.CliHomeFlag))

	cliCtx, err := cliContext.NewCLIContext(
		viper.GetString(common.ChainIDFlag),
		viper.GetString(common.NodeEndpointFlag),
		viper.GetString(common.UserNameFlag),
		viper.GetBool(common.GenOnlyFlag),
		viper.GetString(common.BroadcastModeFlag),
		viper.GetString(common.VfrHomeFlag),
		viper.GetInt64(common.HeightFlag),
		viper.GetBool(common.TrustNodeFlag),
		viper.GetString(common.CliHomeFlag),
		"")
	if err != nil {
		return cliContext.CLIContext{}, nil, err
	}
	cliCtx = cliCtx.WithCodec(cdc)

	return cliCtx, auth.DefaultTxDecoder(cdc), nil
}
