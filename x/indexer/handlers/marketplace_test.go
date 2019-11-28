package handlers

import (
	"os/user"
	"path"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	cliContext "github.com/dgamingfoundation/cosmos-utils/client/context"
	common "github.com/dgamingfoundation/dwh/x/common"
	app "github.com/dgamingfoundation/marketplace"
	"github.com/stretchr/testify/require"
)

func TestEnsureUserExists(t *testing.T) {
	cfg := common.DefaultDwhCommonServiceConfig()
	db, err := common.GetDB(cfg)
	if err != nil {
		t.Errorf("failed to establish database connection: %v", err)
		return
	}
	var (
		sender, _    = sdk.AccAddressFromHex("cosmos1tctr64k4en25uvet2k2tfkwkh0geyrv8fvuvet")
		recipient, _ = sdk.AccAddressFromHex("cosmos1tctr64k4en25uvet2k2tfkwkh0geyrv8fvuvet")
	)

	msgMintNFT := nft.MsgMintNFT{
		Sender:    sender,
		Recipient: recipient,
	}
	sdkMsg := sdk.Msg(msgMintNFT)
	handler := &MarketplaceHandler{}
	addresses, err := handler.getMsgAddresses(db, sdkMsg)
	require.NoError(t, err)

	require.Equal(t, 2, len(addresses))
}

func TestMarketplaceHandlerResetAndSetup(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Errorf("failed to get current user, exiting: %v", err)
		return
	}

	cfg := common.DefaultDwhCommonServiceConfig()
	db, err := common.GetDB(cfg)
	if err != nil {
		t.Errorf("failed to establish database connection: %v", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database connection: %v", err)
		}
	}()
	cdc := app.MakeCodec()

	cliCtx, err := cliContext.NewContext(
		"mpchain",
		"tcp://localhost:26657",
		path.Join(usr.HomeDir, ".mpcli"))
	if err != nil {
		t.Errorf("failed to create cliCtx connection: %v", err)
		return
	}
	cliCtx = cliCtx.WithCodec(cdc)

	handler := NewMarketplaceHandler(cliCtx)
	db, err = handler.Reset(db)
	if err != nil {
		t.Errorf("failed to Reset db: %v", err)
	}
	require.False(t, db.HasTable(&common.NFT{}))
	require.False(t, db.HasTable(&common.FungibleToken{}))
	require.False(t, db.HasTable(&common.FungibleTokenTransfer{}))
	require.False(t, db.HasTable(&common.User{}))

	db, err = handler.Setup(db)
	if err != nil {
		t.Errorf("failed to Setup db: %v", err)
	}

	require.True(t, db.HasTable(&common.NFT{}))
	require.True(t, db.HasTable(&common.FungibleToken{}))
	require.True(t, db.HasTable(&common.FungibleTokenTransfer{}))
	require.True(t, db.HasTable(&common.User{}))

	db, err = handler.Reset(db)
	if err != nil {
		t.Errorf("failed to Reset db: %v", err)
	}

	require.False(t, db.HasTable(&common.NFT{}))
	require.False(t, db.HasTable(&common.FungibleToken{}))
	require.False(t, db.HasTable(&common.FungibleTokenTransfer{}))
	require.False(t, db.HasTable(&common.User{}))
}
