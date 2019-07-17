package indexer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/prometheus/common/log"
)

type MsgHandler func(msg sdk.Msg) error

func NewMarketplaceHandler() MsgHandler {
	return func(msg sdk.Msg) error {
		switch value := msg.(type) {
		case types.MsgMintNFT:
			log.Infof("got message of type MsgMintNFT: %+v", value)
		case types.MsgSellNFT:
			log.Infof("got message of type MsgSellNFT: %+v", value)
		case types.MsgBuyNFT:
			log.Infof("got message of type MsgBuyNFT: %+v", value)
		case types.MsgTransferNFT:
			log.Infof("got message of type MsgTransferNFT: %+v", value)
		}

		return nil
	}
}
