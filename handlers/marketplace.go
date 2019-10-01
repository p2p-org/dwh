package handlers

import (
	"database/sql"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	cliContext "github.com/dgamingfoundation/cosmos-utils/client/context"
	"github.com/dgamingfoundation/dwh/common"
	app "github.com/dgamingfoundation/marketplace"
	mptypes "github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/common/log"
	"github.com/tendermint/go-amino"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type MarketplaceHandler struct {
	cdc        *amino.Codec
	cliCtx     cliContext.Context
	msgMetrics *common.MsgMetrics
}

func NewMarketplaceHandler(cliCtx cliContext.Context) MsgHandler {
	msgMetr := common.NewPrometheusMsgMetrics("marketplace")
	return &MarketplaceHandler{
		cdc:        app.MakeCodec(),
		cliCtx:     cliCtx,
		msgMetrics: msgMetr,
	}
}

func (m *MarketplaceHandler) findOrCreateUser(db *gorm.DB, accAddress sdk.AccAddress) (*common.User, error) {
	user := &common.User{}
	acc, err := m.getAccount(accAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to find owner with addr %s: %v", accAddress.String(), err)
	}
	row := db.Table("users").Where("address = ?", accAddress.String()).Row()
	err = row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.Name, &user.Address, &user.Balance, &user.AccountNumber, &user.SequenceNumber)
	if err == sql.ErrNoRows {
		// Create a new user.
		user = common.NewUser(
			"",
			acc.GetAddress(),
			acc.GetCoins(),
			acc.GetAccountNumber(),
			acc.GetSequence(),
			nil,
		)
		if db.Create(&user).Error != nil {
			return nil, fmt.Errorf("failed to add user for address %s: %v", accAddress, db.Error)
		}
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	user.SequenceNumber = acc.GetSequence()
	db = db.Model(&user).Update("sequence_number", user.SequenceNumber)
	if db.Error != nil {
		return nil, fmt.Errorf("failed to update user %s: %v", accAddress, db.Error)
	}
	return user, nil
}

func (m *MarketplaceHandler) increaseCounter(labels ...string) {
	counter, err := m.msgMetrics.NumMsgs.GetMetricWithLabelValues(labels...)
	if err != nil {
		log.Errorf("get metrics with label values error: %v", err)
		return
	}
	counter.Inc()
}

func (m *MarketplaceHandler) Handle(db *gorm.DB, msg sdk.Msg, events ...*abciTypes.Event) error {
	m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueCommon)
	switch value := msg.(type) {
	case nft.MsgMintNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgMintNFT)
		if _, err := m.findOrCreateUser(db, value.Recipient); err != nil {
			return err
		}
		log.Infof("got message of type MsgMintNFT: %+v", value)
		token, err := m.getNFT(value.ID)
		if err != nil {
			return fmt.Errorf("failed to getNFT: %v", err)
		}

		db = db.Create(token)
		if db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgMintNFT)
	case mptypes.MsgPutNFTOnMarket:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgPutNFTOnMarket)
		log.Infof("got message of type MsgSellNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":            true,
			"Price":             value.Price.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgSellNFT): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgPutNFTOnMarket)
	case mptypes.MsgBuyNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgBuyNFT)
		log.Infof("got message of type MsgBuyNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Buyer.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgBuyNFT): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgBuyNFT)
	case nft.MsgTransferNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgTransferNFT)
		log.Infof("got message of type MsgTransferNFT: %+v", value)
		db = db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Recipient.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgTransferNFT): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgTransferNFT)
		// Also: MsgDeleteNFT (not implemented yet).
	case mptypes.MsgCreateFungibleToken:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgCreateFungibleToken)
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
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgCreateFungibleToken)
	case mptypes.MsgTransferFungibleTokens:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgTransferFungibleTokens)
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
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgTransferFungibleTokens)
	}
	m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueCommon)
	return nil
}

func (m *MarketplaceHandler) RouterKeys() []string {
	return []string{mptypes.ModuleName, nft.ModuleName}
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

	var token mptypes.NFTInfo
	if err := m.cdc.UnmarshalJSON(res, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NFT: %v", err)
	}

	return common.NewNFTFromMarketplaceNFT(&token), nil
}

func (m *MarketplaceHandler) getAccount(addr sdk.AccAddress) (exported.Account, error) {
	accGetter := authtypes.NewAccountRetriever(m.cliCtx)
	if err := accGetter.EnsureExists(addr); err != nil {
		return nil, err
	}

	return accGetter.GetAccount(addr)
}
