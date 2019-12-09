package handlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cliContext "github.com/corestario/cosmos-utils/client/context"
	common "github.com/corestario/dwh/x/common"
	app "github.com/corestario/marketplace"
)

func GetEnv(config *common.DwhCommonServiceConfig) (cliContext.Context, sdk.TxDecoder, error) {
	cdc := app.MakeCodec()
	cliCtx, err := cliContext.NewContext(
		config.ChainID,
		config.MarketplaceAddr,
		config.CliHome,
	)
	if err != nil {
		return cliContext.Context{}, nil, err
	}
	cliCtx = cliCtx.WithCodec(cdc)

	return cliCtx, auth.DefaultTxDecoder(cdc), nil
}
