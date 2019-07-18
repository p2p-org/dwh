package indexer

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/dwh/common"
	app "github.com/dgamingfoundation/marketplace"
	mptypes "github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/common/log"
	"github.com/tendermint/go-amino"
)

type MsgHandler interface {
	Handle(msg sdk.Msg) error
}

type MarketplaceHandler struct {
	db     *gorm.DB
	cdc    *amino.Codec
	cliCtx client.CLIContext
}

func NewMarketplaceHandler(db *gorm.DB, cliCtx client.CLIContext) MsgHandler {
	return &MarketplaceHandler{
		db:     db,
		cdc:    app.MakeCodec(),
		cliCtx: cliCtx,
	}
}

func (m *MarketplaceHandler) Handle(msg sdk.Msg) error {
	switch value := msg.(type) {
	case mptypes.MsgMintNFT:
		log.Infof("got message of type MsgMintNFT: %+v", value)
		res, _, err := m.cliCtx.QueryWithData(fmt.Sprintf("custom/marketplace/nft/%s", value.UUID), nil)
		if err != nil {
			return fmt.Errorf("could not find token with UUID %s: %v", value.UUID, err)
		}

		var nft mptypes.NFT
		if err := m.cdc.UnmarshalJSON(res, &nft); err != nil {
			return fmt.Errorf("failed to unmarshal NFT: %v", err)
		}

		m.db.Create(common.NewNFTFromMarketplaceNFT(&nft))
	case mptypes.MsgSellNFT:
		log.Infof("got message of type MsgSellNFT: %+v", value)
	case mptypes.MsgBuyNFT:
		log.Infof("got message of type MsgBuyNFT: %+v", value)
	case mptypes.MsgTransferNFT:
		log.Infof("got message of type MsgTransferNFT: %+v", value)
		// Also: MsgDeleteNFT (not implemented yet).
	}

	return nil
}
