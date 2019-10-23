package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/dgamingfoundation/dwh/x/indexer/handlers"
	app "github.com/dgamingfoundation/marketplace"
)

func TestMarketplaceHandlerResetAndSetup(t *testing.T) {
	db, err := common.GetDB()
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
	ctx := cliContext.NewCLIContext().WithCodec(cdc)
	handler := handlers.NewMarketplaceHandler(ctx)
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
