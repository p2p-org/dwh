package indexer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	cliContext "github.com/dgamingfoundation/dkglib/lib/client/context"
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
	cliCtx cliContext.CLIContext
}

func NewMarketplaceHandler(db *gorm.DB, cliCtx cliContext.CLIContext) MsgHandler {
	return &MarketplaceHandler{
		db:     db,
		cdc:    app.MakeCodec(),
		cliCtx: cliCtx,
	}
}

func (m *MarketplaceHandler) findOrCreateUser(accAddress sdk.AccAddress) (*common.User, error) {
	user := &common.User{}
	m.db.Where("Address = ?", accAddress.String()).First(&user)
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
		m.db = m.db.Create(&user)
		if m.db.Error != nil {
			return nil, fmt.Errorf("failed to add user for address %s: %v", accAddress, m.db.Error)
		}
	}
	return user, nil
}

func (m *MarketplaceHandler) Handle(msg sdk.Msg) error {
	switch value := msg.(type) {
	case mptypes.MsgMintNFT:
		if _, err := m.findOrCreateUser(value.Owner); err != nil {
			return err
		}
		log.Infof("got message of type MsgMintNFT: %+v", value)
		m.db = m.db.Create(&common.NFT{
			OwnerAddress: value.Owner.String(),
			TokenID:      value.TokenID,
			Name:         value.Name,
			Description:  value.Description,
			Image:        value.Image,
			TokenURI:     value.TokenURI,
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", m.db.Error)
		}
	case mptypes.MsgPutNFTOnMarket:
		log.Infof("got message of type MsgSellNFT: %+v", value)
		m.db = m.db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":            true,
			"Price":             value.Price.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgSellNFT): %v", m.db.Error)
		}
	case mptypes.MsgBuyNFT:
		log.Infof("got message of type MsgBuyNFT: %+v", value)
		m.db = m.db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Buyer.String(),
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgBuyNFT): %v", m.db.Error)
		}
	case mptypes.MsgTransferNFT:
		log.Infof("got message of type MsgTransferNFT: %+v", value)
		m.db = m.db.Model(&common.NFT{}).UpdateColumns(map[string]interface{}{
			"OnSale":       false,
			"OwnerAddress": value.Recipient.String(),
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgTransferNFT): %v", m.db.Error)
		}
		// Also: MsgDeleteNFT (not implemented yet).
	case mptypes.MsgCreateFungibleToken:
		log.Infof("got message of type MsgCreateFungibleToken: %+v", value)
		if _, err := m.findOrCreateUser(value.Creator); err != nil {
			return err
		}
		m.db = m.db.Create(&common.FungibleToken{
			OwnerAddress:   value.Creator.String(),
			Denom:          value.Denom,
			EmissionAmount: value.Amount,
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", m.db.Error)
		}
	case mptypes.MsgTransferFungibleTokens:
		log.Infof("got message of type MsgTransferFungibleTokens: %+v", value)
		var (
			ft                common.FungibleToken
			sender, recipient *common.User
			err               error
		)
		if sender, err = m.findOrCreateUser(value.Owner); err != nil {
			return err
		}
		if recipient, err = m.findOrCreateUser(value.Recipient); err != nil {
			return err
		}
		m.db.Where("denom = ?", value.Denom).First(&ft)
		if ft.ID == 0 {
			return fmt.Errorf("failed to transfer fungible token: %v", m.db.Error)
		}
		m.db.Model(&ft).Association("FungibleTokenTransfers").Append(common.FungibleTokenTransfer{
			SenderAddress:    sender.Address,
			RecipientAddress: recipient.Address,
			Amount:           value.Amount,
		})
		if m.db.Error != nil {
			return fmt.Errorf("failed to transfer fungible token: %v", m.db.Error)
		}
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

func (m *MarketplaceHandler) getAccount(addr sdk.AccAddress) (exported.Account, error) {
	accGetter := authtypes.NewAccountRetriever(m.cliCtx)
	if err := accGetter.EnsureExists(addr); err != nil {
		return nil, err
	}

	return accGetter.GetAccount(addr)
}
