package common

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
)

type NFT struct {
	gorm.Model
	OwnerAddress      string
	TokenID           string
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
		TokenID:           nft.GetID(),
		OwnerAddress:      nft.GetOwner().String(),
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

type User struct {
	gorm.Model
	Name    string
	Address string `gorm:"unique;not null"`
	Balance string
	Tokens  []NFT `gorm:"ForeignKey:OwnerAddress"`
}

func NewUser(name string, addr sdk.AccAddress, balance sdk.Coins, tokens []*NFT) *User {
	return &User{
		Name:    name,
		Address: addr.String(),
		Balance: balance.String(),
		Tokens:  tokens,
	}
}
