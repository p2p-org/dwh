package dwh_common

import (
	"encoding/json"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

type ImgQueuePriority uint8

const (
	RegularUpdatePriority ImgQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

type Resolution struct {
	Width  uint `mapstructure:"width",json:"width"`
	Height uint `mapstructure:"height",json:"height"`
}

type TaskInfo struct {
	Owner   string `json:"owner"`
	TokenID string `json:"token_id"`
	URL     string `json:"url"`
}

type NFT struct {
	gorm.Model
	Denom             string
	TokenID           string `gorm:"unique;not null"`
	OwnerAddress      string `gorm:"type:varchar(45)"`
	TokenURI          string
	Status            int
	Price             string
	SellerBeneficiary string

	// Auction-related fields
	BuyoutPrice  string
	OpeningPrice string
	TimeToSell   time.Time

	// Relations
	Offers []Offer      `gorm:"ForeignKey:TokenID"`
	Bids   []AuctionBid `gorm:"ForeignKey:TokenID"`
}

func NewNFTFromMarketplaceNFT(denom, tokenID, ownerAddress, tokenURI string) *NFT {
	return &NFT{
		Denom:        denom,
		TokenID:      tokenID,
		OwnerAddress: ownerAddress,
		TokenURI:     tokenURI,
		Status:       int(types.NFTStatusDefault),
	}
}

type Offer struct {
	gorm.Model
	OfferID               string
	Buyer                 string
	Price                 string
	BuyerBeneficiary      string
	BeneficiaryCommission string
	TokenID               string
}

func NewOffer(offer *types.Offer, tokenID string) *Offer {
	return &Offer{
		OfferID:               offer.ID,
		Buyer:                 offer.Buyer.String(),
		BuyerBeneficiary:      offer.BuyerBeneficiary.String(),
		BeneficiaryCommission: offer.BeneficiaryCommission,
		TokenID:               tokenID,
		Price:                 offer.Price.String(),
	}
}

type AuctionBid struct {
	gorm.Model
	BidderAddress         string
	BidderBeneficiary     string
	BeneficiaryCommission string
	Price                 string
	TokenID               string
}

type FungibleToken struct {
	gorm.Model
	OwnerAddress           string `gorm:"type:varchar(45)"`
	Denom                  string `gorm:"unique;not null"`
	EmissionAmount         int64
	FungibleTokenTransfers []FungibleTokenTransfer `gorm:"ForeignKey:FungibleTokenID"`
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
	AccountNumber  uint64
	SequenceNumber uint64
	Tokens         []NFT           `gorm:"ForeignKey:OwnerAddress"`
	FungibleTokens []FungibleToken `gorm:"ForeignKey:OwnerAddress"`
}

func NewUser(name string, addr sdk.AccAddress, balance sdk.Coins, accountNumber, sequenceNumber uint64, tokens []NFT) *User {
	return &User{
		Name:           name,
		Address:        addr.String(),
		Balance:        balance.String(),
		AccountNumber:  accountNumber,
		SequenceNumber: sequenceNumber,
		Tokens:         tokens,
	}
}

type Tx struct {
	gorm.Model
	Hash      string `gorm:"not null"`
	Height    int64  `gorm:"not null"`
	Index     uint32 `gorm:"not null"`
	Code      uint32 `gorm:"not null"`
	Data      []byte
	Log       postgres.Jsonb
	Info      string
	GasWanted int64
	GasUsed   int64
	Messages  []Message `gorm:"ForeignKey:TxIndex"`
}

func NewTx(tx *coreTypes.ResultTx) *Tx {
	return &Tx{
		Hash:      tx.Hash.String(),
		Height:    tx.Height,
		Index:     tx.Index,
		Code:      tx.TxResult.Code,
		Data:      tx.TxResult.Data,
		Log:       postgres.Jsonb{json.RawMessage(tx.TxResult.Log)},
		Info:      tx.TxResult.Info,
		GasWanted: tx.TxResult.GasWanted,
		GasUsed:   tx.TxResult.GasUsed,
	}
}

type Message struct {
	gorm.Model
	Route     string
	MsgType   string
	Signature postgres.Jsonb
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
		Signature: postgres.Jsonb{json.RawMessage(sign)},
		Signers:   strings.Join(strSigners, ", "),
		Failed:    failed,
		Error:     error,
		TxID:      txID,
	}
}
