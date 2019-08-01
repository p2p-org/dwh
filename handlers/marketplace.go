package handlers

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dgamingfoundation/dwh/common"
	app "github.com/dgamingfoundation/marketplace"
	mptypes "github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/common/log"
	"github.com/tendermint/go-amino"
)

type MarketplaceHandler struct {
	cdc        *amino.Codec
	cliCtx     client.CLIContext
	msgMetrics *common.MsgMetrics
}

func NewMarketplaceHandler(cliCtx client.CLIContext) MsgHandler {
	msgMetr := common.NewPrometheusMsgMetrics("marketplace")
	return &MarketplaceHandler{
		cdc:        app.MakeCodec(),
		cliCtx:     cliCtx,
		msgMetrics: msgMetr,
	}
}

func (m *MarketplaceHandler) findOrCreateUser(db *gorm.DB, accAddress sdk.AccAddress) (*common.User, error) {
	user := &common.User{}
	db.Where("Address = ?", accAddress.String()).First(&user)
	if len(user.Address) == 0 {
		// Create a new user.
		acc, err := m.getAccount(accAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to find owner with addr %s: %v", accAddress.String(), err)
		}
		user = common.NewUser(
			"",
			acc.GetAddress(),
			acc.GetCoins(),
			nil,
		)
		db = db.Create(&user)
		if db.Error != nil {
			return nil, fmt.Errorf("failed to add user for address %s: %v", accAddress, db.Error)
		}
	}
	return user, nil
}

func (m *MarketplaceHandler) Handle(db *gorm.DB, msg sdk.Msg) error {
	m.msgMetrics.NumMsgs.Inc()
	switch value := msg.(type) {
	case mptypes.MsgMintNFT:
		if _, err := m.findOrCreateUser(db, value.Owner); err != nil {
			return err
		}
		log.Infof("got message of type MsgMintNFT: %+v", value)
		db = db.Create(&common.NFT{
			OwnerAddress: value.Owner.String(),
			TokenID:      value.TokenID,
			Name:         value.Name,
			Description:  value.Description,
			Image:        value.Image,
			TokenURI:     value.TokenURI,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", db.Error)
		}
	case mptypes.MsgPutNFTOnMarket:
		log.Infof("got message of type MsgSellNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":            true,
			"Price":             value.Price.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgSellNFT): %v", db.Error)
		}
	case mptypes.MsgBuyNFT:
		log.Infof("got message of type MsgBuyNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Buyer.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgBuyNFT): %v", db.Error)
		}
	case mptypes.MsgTransferNFT:
		log.Infof("got message of type MsgTransferNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Recipient.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgTransferNFT): %v", db.Error)
		}
		// Also: MsgDeleteNFT (not implemented yet).
	case mptypes.MsgCreateFungibleToken:
		log.Infof("got message of type MsgCreateFungibleToken: %+v", value)
		if _, err := m.findOrCreateUser(db, value.Creator); err != nil {
			return err
		}
		db = db.Create(&common.FungibleToken{
			OwnerAddress:   value.Creator.String(),
			Denom:          value.Denom,
			EmissionAmount: value.Amount,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", db.Error)
		}
	case mptypes.MsgTransferFungibleTokens:
		log.Infof("got message of type MsgTransferFungibleTokens: %+v", value)
		var (
			ft                common.FungibleToken
			sender, recipient *common.User
			err               error
		)
		if sender, err = m.findOrCreateUser(db, value.Owner); err != nil {
			return err
		}
		if recipient, err = m.findOrCreateUser(db, value.Recipient); err != nil {
			return err
		}
		db.Where("denom = ?", value.Denom).First(&ft)
		if ft.ID == 0 {
			return fmt.Errorf("failed to transfer fungible token: %v", db.Error)
		}
		db.Model(&ft).Association("FungibleTokenTransfers").Append(common.FungibleTokenTransfer{
			SenderAddress:    sender.Address,
			RecipientAddress: recipient.Address,
			Amount:           value.Amount,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to transfer fungible token: %v", db.Error)
		}
	}
	m.msgMetrics.NumMsgsAccepted.Inc()
	return nil
}

func (m *MarketplaceHandler) RouterKey() string {
	return mptypes.ModuleName
}

func (m *MarketplaceHandler) Setup(db *gorm.DB) (*gorm.DB, error) {
	db = db.CreateTable(&common.NFT{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table nfts: %v", db.Error)
	}
	db = db.CreateTable(&common.FungibleToken{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table fungible_tokens: %v", db.Error)
	}
	db = db.CreateTable(&common.FungibleTokenTransfer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table fungible_token_transfers: %v", db.Error)
	}
	db = db.CreateTable(&common.User{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table users: %v", db.Error)
	}

	db = db.Model(&common.NFT{}).AddForeignKey(
		"owner_address", "users(address)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (nfts): %v", db.Error)
	}
	db = db.Model(&common.FungibleToken{}).AddForeignKey(
		"owner_address", "users(address)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (fuingible_tokens): %v", db.Error)
	}

	db = db.Model(&common.FungibleTokenTransfer{}).AddForeignKey(
		"sender_address", "users(address)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (fungible_tokens_transfers): %v", db.Error)
	}
	db = db.Model(&common.FungibleTokenTransfer{}).AddForeignKey(
		"recipient_address", "users(address)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (fungible_tokens_transfers): %v", db.Error)
	}
	db = db.Model(&common.FungibleTokenTransfer{}).AddForeignKey(
		"fungible_token_id", "fungible_tokens(id)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (fungible_tokens_transfers): %v", db.Error)
	}

	return db, nil
}

func (m *MarketplaceHandler) Reset(db *gorm.DB) (*gorm.DB, error) {
	db = db.DropTableIfExists(&common.NFT{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table nfts: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.FungibleTokenTransfer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table fungible_tokens: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.FungibleToken{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table fungible_tokens: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.User{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table users: %v", db.Error)
	}

	return db, nil
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

func (m *MarketplaceHandler) getAccount(addr sdk.AccAddress) (exported.Account, error) {
	accGetter := authtypes.NewAccountRetriever(m.cliCtx)
	if err := accGetter.EnsureExists(addr); err != nil {
		return nil, err
	}

	return accGetter.GetAccount(addr)
}