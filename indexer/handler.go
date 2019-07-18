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
		m.db.Create(&common.NFT{
			Owner:       value.Owner.String(),
			TokenID:     value.TokenID,
			Name:        value.Name,
			Description: value.Description,
			Image:       value.Image,
			TokenURI:    value.TokenURI,
		})
	case mptypes.MsgSellNFT:
		log.Infof("got message of type MsgSellNFT: %+v", value)
		m.db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":            true,
			"Price":             value.Price.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
		})
	case mptypes.MsgBuyNFT:
		log.Infof("got message of type MsgBuyNFT: %+v", value)
	case mptypes.MsgTransferNFT:
		log.Infof("got message of type MsgTransferNFT: %+v", value)
		// Also: MsgDeleteNFT (not implemented yet).
	}

	return nil
}

func (m *MarketplaceHandler) getNFT(tokenID string) (*common.NFT, error) {
	res, _, err := m.cliCtx.QueryWithData(fmt.Sprintf("custom/marketplace/nft/%s", tokenID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not find token with TokenID %s: %v", tokenID, err)
	}

	var nft mptypes.NFT
	if err := m.cdc.UnmarshalJSON(res, &nft); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NFT: %v", err)
	}

	return common.NewNFTFromMarketplaceNFT(&nft), nil
}
