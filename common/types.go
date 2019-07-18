package common

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
)

type NFT struct {
	gorm.Model
	UUID              string
	Name              string
	Description       string
	Image             string
	TokenURI          string
	OnSale            bool
	Price             string
	SellerBeneficiary string
}

func NewNFTFromMarketplaceNFT(nft *types.NFT) *NFT {
	return &NFT{
		UUID:              nft.GetID(),
		Name:              nft.GetName(),
		Description:       nft.GetDescription(),
		Image:             nft.GetImage(),
		TokenURI:          nft.GetTokenURI(),
		OnSale:            nft.IsOnSale(),
		Price:             nft.GetPrice().String(),
		SellerBeneficiary: nft.SellerBeneficiary.String(),
	}
}

type Message struct {
	gorm.Model
	Height    int64
	TxHash    string
	Route     string
	MsgType   string
	Signature string
	Signers   string
	Failed    bool
	Error     string
}

func NewMessage(
	height int64,
	txHash string,
	route,
	msgType string,
	sign string,
	signers []sdk.AccAddress,
	failed bool,
	error string,
) *Message {
	var strSigners []string
	for _, signer := range signers {
		strSigners = append(strSigners, signer.String())
	}

	return &Message{
		Height:    height,
		TxHash:    txHash,
		Route:     route,
		MsgType:   msgType,
		Signature: sign,
		Signers:   strings.Join(strSigners, ", "),
		Failed:    failed,
		Error:     error,
	}
}
