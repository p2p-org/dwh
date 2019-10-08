package handlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cliContext "github.com/dgamingfoundation/cosmos-utils/client/context"
	"github.com/dgamingfoundation/dwh/common"
	app "github.com/dgamingfoundation/marketplace"
	"github.com/spf13/viper"
)

func GetEnv() (cliContext.Context, sdk.TxDecoder, error) {
	cdc := app.MakeCodec()
	cliCtx, err := cliContext.NewContext(
		viper.GetString(common.ChainIDFlag),
		viper.GetString(common.NodeEndpointFlag),
		viper.GetString(common.CliHomeFlag))
	if err != nil {
		return cliContext.Context{}, nil, err
	}
	cliCtx = cliCtx.WithCodec(cdc)

	return cliCtx, auth.DefaultTxDecoder(cdc), nil
}
