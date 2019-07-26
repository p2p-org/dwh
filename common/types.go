package common

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

// TODO: add normal types for fields in models (varchar, etc.). Issue #19 in the marketplace may be useful

type NFT struct {
	gorm.Model
	OwnerAddress      string `gorm:"type:varchar(45)"`
	TokenID           string `gorm:"unique;not null"`
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

type FungibleToken struct {
	gorm.Model
	OwnerAddress           string `gorm:"type:varchar(45)"`
	Denom                  string `gorm:"unique;not null"`
	EmissionAmount         int64
	FungibleTokenTransfers []FungibleTokenTransfer `gorm:"ForeignKey:FungibleTokenID"`
}

func NewFungibleTokenFromMarketplace(ft *types.FungibleToken) *FungibleToken {
	return &FungibleToken{
		OwnerAddress:   ft.Creator.String(),
		Denom:          ft.Denom,
		EmissionAmount: ft.EmissionAmount,
	}
}

type FungibleTokenTransfer struct {
	gorm.Model
	SenderAddress    string `gorm:"type:varchar(45)"`
	RecipientAddress string `gorm:"type:varchar(45)"`
	FungibleTokenID  int64
	Amount           int64
}

type User struct {
	gorm.Model
	Name           string
	Address        string `gorm:"type:varchar(45);unique;not null"`
	Balance        string
	Tokens         []NFT           `gorm:"ForeignKey:OwnerAddress"`
	FungibleTokens []FungibleToken `gorm:"ForeignKey:OwnerAddress"`
}

func NewUser(name string, addr sdk.AccAddress, balance sdk.Coins, tokens []NFT) *User {
	return &User{
		Name:    name,
		Address: addr.String(),
		Balance: balance.String(),
		Tokens:  tokens,
	}
}

type Tx struct {
	gorm.Model
	Hash      string `gorm:"not null"`
	Height    int64  `gorm:"not null"`
	Index     uint32 `gorm:"not null"`
	Code      uint32 `gorm:"not null"`
	Data      []byte
	Log       string
	Info      string
	GasWanted int64
	GasUsed   int64
	Messages  []Message `gorm:"ForeignKey:TxIndex"`
}

func NewTx(tx *core_types.ResultTx) *Tx {
	return &Tx{
		Hash:      tx.Hash.String(),
		Height:    tx.Height,
		Index:     tx.Index,
		Code:      tx.TxResult.Code,
		Data:      tx.TxResult.Data,
		Log:       tx.TxResult.Log,
		Info:      tx.TxResult.Info,
		GasWanted: tx.TxResult.GasWanted,
		GasUsed:   tx.TxResult.GasUsed,
	}
}

type Message struct {
	gorm.Model
	Route     string
	MsgType   string
	Signature string
	Signers   string
	Failed    bool
	Error     string
	TxID      uint
}

func NewMessage(
	route,
	msgType string,
	sign string,
	signers []sdk.AccAddress,
	failed bool,
	error string,
	txID uint,
) *Message {
	var strSigners []string
	for _, signer := range signers {
		strSigners = append(strSigners, signer.String())
	}

	return &Message{
		Route:     route,
		MsgType:   msgType,
		Signature: sign,
		Signers:   strings.Join(strSigners, ", "),
		Failed:    failed,
		Error:     error,
		TxID:      txID,
	}
}
