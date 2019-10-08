package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	cliContext "github.com/dgamingfoundation/cosmos-utils/client/context"
	"github.com/dgamingfoundation/dwh/common"
	dwh_common "github.com/dgamingfoundation/dwh/x/common"
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
	uriSender  *dwh_common.RMQSender
}

func NewMarketplaceHandler(cliCtx cliContext.Context) MsgHandler {
	msgMetr := common.NewPrometheusMsgMetrics("marketplace")
	cfg := dwh_common.ReadCommonConfig(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)

	sender, err := dwh_common.NewRMQSender(cfg, cfg.UriQueueName, cfg.UriQueueMaxPriority)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return &MarketplaceHandler{
		cdc:        app.MakeCodec(),
		cliCtx:     cliCtx,
		msgMetrics: msgMetr,
		uriSender:  sender,
	}
}

func (m *MarketplaceHandler) findOrCreateUser(db *gorm.DB, accAddress sdk.AccAddress) (*common.User, error) {
	user := &common.User{}
	acc, err := m.getAccount(accAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to find owner with addr %s: %v", accAddress.String(), err)
	}
	row := db.Table("users").Where("address = ?", accAddress.String()).Row()
	err = row.Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Name,
		&user.Address,
		&user.Balance,
		&user.AccountNumber,
		&user.SequenceNumber)
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

func (m *MarketplaceHandler) Handle(db *gorm.DB, msg sdk.Msg, events ...abciTypes.Event) error {
	m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueCommon)
	log.Infof("got message of type %s: %+v", msg.Type(), msg)
	switch value := msg.(type) {
	case nft.MsgMintNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgMintNFT)
		if _, err := m.findOrCreateUser(db, value.Recipient); err != nil {
			return err
		}
		db = db.Create(
			common.NewNFTFromMarketplaceNFT(value.ID, value.Recipient.String(), value.TokenURI),
		)
		if db.Error != nil {
			return fmt.Errorf("failed to create nft: %v", db.Error)
		}
		if err := m.uriSender.Publish(value.TokenURI, value.Recipient.String(), value.ID, dwh_common.FreshlyMadePriority); err != nil {
			return fmt.Errorf("failed to send message to RabbitMQ: %v", err)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgMintNFT)
	case nft.MsgEditNFTMetadata:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgEditNFTMetadata)
		if _, err := m.findOrCreateUser(db, value.Sender); err != nil {
			return err
		}
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.ID).UpdateColumn(map[string]interface{}{
			"TokenURI": value.TokenURI,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgEditNFTMetadata): %v", db.Error)
		}
		if err := m.uriSender.Publish(value.TokenURI, value.Sender.String(), value.ID, dwh_common.ForcedUpdatesPriority); err != nil {
			return fmt.Errorf("failed to send message to RabbitMQ: %v", err)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgEditNFTMetadata)
	case nft.MsgTransferNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgTransferNFT)
		if _, err := m.findOrCreateUser(db, value.Recipient); err != nil {
			return err
		}
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.ID).UpdateColumns(map[string]interface{}{
			"OwnerAddress": value.Recipient.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgTransferNFT): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgTransferNFT)
	case mptypes.MsgPutNFTOnMarket:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgPutNFTOnMarket)
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"Status":            mptypes.NFTStatusOnMarket,
			"Price":             value.Price.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgPutNFTOnMarket): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgPutNFTOnMarket)
	case mptypes.MsgRemoveNFTFromMarket:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgRemoveNFTFromMarket)
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"Status":            mptypes.NFTStatusDefault,
			"SellerBeneficiary": "",
			"Price":             sdk.Coins{}.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgRemoveNFTFromMarket): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgRemoveNFTFromMarket)
	case mptypes.MsgBuyNFT:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgBuyNFT)
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"Status":       mptypes.NFTStatusDefault,
			"OwnerAddress": value.Buyer.String(),
			"Price":        sdk.Coins{}.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgBuyNFT): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgBuyNFT)
	case mptypes.MsgPutNFTOnAuction:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgPutNFTOnAuction)
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"Status":            mptypes.NFTStatusOnAuction,
			"BuyoutPrice":       value.BuyoutPrice.String(),
			"OpeningPrice":      value.OpeningPrice.String(),
			"SellerBeneficiary": value.Beneficiary.String(),
			"TimeToSell":        value.TimeToSell.String(),
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgPutNFTOnAuction): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgPutNFTOnAuction)
	case mptypes.MsgRemoveNFTFromAuction:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgRemoveFromAuction)
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"Status":            mptypes.NFTStatusDefault,
			"BuyoutPrice":       sdk.Coins{}.String(),
			"OpeningPrice":      sdk.Coins{}.String(),
			"SellerBeneficiary": "",
			"TimeToSell":        0,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgRemoveNFTFromAuction): %v", db.Error)
		}
		db.Where("token_id = ?", value.TokenID).Delete(&common.AuctionBid{})
		if db.Error != nil {
			return fmt.Errorf("failed to delete bids (MsgRemoveNFTFromAuction): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgRemoveFromAuction)
	case mptypes.MsgMakeBidOnAuction:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgMakeBidOnAuction)
		// Find out whether we had a buyout.
		_, isBuyout := m.getEventAttr(events, msg.Type(), mptypes.AttributeKeyIsBuyout)
		if isBuyout {
			// Reset all auction-related fields, delete all related bids.
			db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
				"OwnerAddress":      value.Bidder.String(),
				"Status":            mptypes.NFTStatusDefault,
				"BuyoutPrice":       sdk.Coins{}.String(),
				"OpeningPrice":      sdk.Coins{}.String(),
				"SellerBeneficiary": "",
				"TimeToSell":        0,
			})
			if db.Error != nil {
				return fmt.Errorf("failed to update token (MsgMakeBidOnAuction): %v", db.Error)
			}
			db = db.Where("token_id = ?", value.TokenID).Delete(&common.AuctionBid{})
			if db.Error != nil {
				return fmt.Errorf("failed to delete auction bids (MsgMakeBidOnAuction): %v", db.Error)
			}
		} else {
			db = db.Create(&common.AuctionBid{
				BidderAddress:         value.Bidder.String(),
				BidderBeneficiary:     value.BuyerBeneficiary.String(),
				BeneficiaryCommission: value.BeneficiaryCommission,
				Price:                 value.Bid.String(),
				TokenID:               value.TokenID,
			})
			if db.Error != nil {
				return fmt.Errorf("failed to add auction bid (MsgMakeBidOnAuction): %v", db.Error)
			}
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgMakeBidOnAuction)
	case mptypes.MsgBuyoutOnAuction:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgBuyoutOnAuction)
		// Reset all auction-related fields, delete all related bids.
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"OwnerAddress":      value.Buyer.String(),
			"Status":            mptypes.NFTStatusDefault,
			"BuyoutPrice":       sdk.Coins{}.String(),
			"OpeningPrice":      sdk.Coins{}.String(),
			"SellerBeneficiary": "",
			"TimeToSell":        0,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to transfer update token (MsgBuyoutOnAuction): %v", db.Error)
		}
		db.Where("token_id = ?", value.TokenID).Delete(&common.AuctionBid{})
		if db.Error != nil {
			return fmt.Errorf("failed to delete auction bids (MsgBuyoutOnAuction): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgBuyoutOnAuction)
	case mptypes.MsgFinishAuction:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgFinishAuction)
		newOwner, ok := m.getEventAttr(events, msg.Type(), mptypes.AttributeKeyOwner)
		if !ok {
			return errors.New("failed to find new owner")
		}
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"OwnerAddress":      newOwner,
			"Status":            mptypes.NFTStatusDefault,
			"BuyoutPrice":       sdk.Coins{}.String(),
			"OpeningPrice":      sdk.Coins{}.String(),
			"SellerBeneficiary": "",
			"TimeToSell":        0,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update nft (MsgFinishAuction): %v", db.Error)
		}
		db.Where("token_id = ?", value.TokenID).Delete(&common.AuctionBid{})
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgFinishAuction)
	case mptypes.MsgMakeOffer:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgMakeOffer)
		var offer = &mptypes.Offer{}
		// Retrieve the event that holds the offer ID (it is generated by the application and
		// can not be retrieved from the transaction message).
		offerID, ok := m.getEventAttr(events, msg.Type(), mptypes.AttributeKeyOfferID)
		if !ok {
			return fmt.Errorf("failed to find offer for token %s", value.TokenID)
		}
		offer.ID = offerID
		offer.Price = value.Price
		offer.Buyer = value.Buyer
		offer.BuyerBeneficiary = value.BuyerBeneficiary
		offer.BeneficiaryCommission = value.BeneficiaryCommission

		db = db.Create(common.NewOffer(offer, value.TokenID))
		if db.Error != nil {
			return fmt.Errorf("failed to create an offer: %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgMakeOffer)
	case mptypes.MsgAcceptOffer:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgAcceptOffer)

		var offers []*common.Offer
		db = db.Where("token_id = ?", value.TokenID).Find(&offers)
		if db.Error != nil {
			return fmt.Errorf("failed to find offers (MsgAcceptOffer): %v", db.Error)
		}
		var offer = &common.Offer{}
		for _, offerCandidate := range offers {
			if offerCandidate.OfferID == value.OfferID {
				*offer = *offerCandidate
			}
		}
		if offer.ID == 0 {
			return fmt.Errorf("unknown offer ID (not found in related offers): %s", value.OfferID)
		}
		db = db.Model(&common.NFT{}).Where("token_id = ?", value.TokenID).UpdateColumns(map[string]interface{}{
			"OwnerAddress": offer.Buyer,
		})
		if db.Error != nil {
			return fmt.Errorf("failed to update token (MsgAcceptOffer): %v", db.Error)
		}
		db.Where("token_id = ?", value.TokenID).Delete(&common.Offer{})
		if db.Error != nil {
			return fmt.Errorf("failed to delete offers (MsgAcceptOffer): %v", db.Error)
		}
		m.increaseCounter(common.PrometheusValueAccepted, common.PrometheusValueMsgAcceptOffer)
	case mptypes.MsgCreateFungibleToken:
		m.increaseCounter(common.PrometheusValueReceived, common.PrometheusValueMsgCreateFungibleToken)
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
		return nil, fmt.Errorf("failed to create table Nfts: %v", db.Error)
	}
	db = db.CreateTable(&common.FungibleToken{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table FungibleTokens: %v", db.Error)
	}
	db = db.CreateTable(&common.FungibleTokenTransfer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table FungibleTokenTransfers: %v", db.Error)
	}
	db = db.CreateTable(&common.User{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table Users: %v", db.Error)
	}
	db = db.CreateTable(&common.Offer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table Offers: %v", db.Error)
	}
	db = db.CreateTable(&common.AuctionBid{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to create table AuctionBids: %v", db.Error)
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

	db = db.Model(&common.Offer{}).AddForeignKey(
		"token_id", "nfts(token_id)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (offers): %v", db.Error)
	}
	db = db.Model(&common.AuctionBid{}).AddForeignKey(
		"token_id", "nfts(token_id)", "CASCADE", "CASCADE")
	if db.Error != nil {
		return nil, fmt.Errorf("failed to add foreign key (auction_bids): %v", db.Error)
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
	db = db.DropTableIfExists(&common.Offer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table Offers: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.AuctionBid{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table AuctionBids: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.NFT{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table Nfts: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.FungibleTokenTransfer{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table FungibleTokenTransferss: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.FungibleToken{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table FungibleTokens: %v", db.Error)
	}
	db = db.DropTableIfExists(&common.User{})
	if db.Error != nil {
		return nil, fmt.Errorf("failed to drop table Users: %v", db.Error)
	}

	return db, nil
}

func (m *MarketplaceHandler) getAccount(addr sdk.AccAddress) (exported.Account, error) {
	accGetter := authtypes.NewAccountRetriever(m.cliCtx)
	if err := accGetter.EnsureExists(addr); err != nil {
		return nil, err
	}

	return accGetter.GetAccount(addr)
}

func (m *MarketplaceHandler) Stop() {
	if err := m.uriSender.Closer(); err != nil {
		fmt.Printf("error occured when stopping indexer marketplaceHandler: %v", err)
	}
}
func (m *MarketplaceHandler) getEventAttr(events []abciTypes.Event, eventType, attrKey string) (string, bool) {
	for _, event := range events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				if string(attr.Key) == attrKey {
					return string(attr.Value), true
				}
			}
		}
	}
	return "", false
}
