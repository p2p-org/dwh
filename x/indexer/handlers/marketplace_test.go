package handlers_test

import (
	"os/user"
	"path"
	"testing"

	cliContext "github.com/dgamingfoundation/cosmos-utils/client/context"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/dgamingfoundation/dwh/x/indexer/handlers"
	app "github.com/dgamingfoundation/marketplace"
	"github.com/stretchr/testify/require"
)

func TestMarketplaceHandlerResetAndSetup(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Errorf("failed to get current user, exiting: %v", err)
		return
	}

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

	cliCtx, err := cliContext.NewContext(
		"mpchain",
		"tcp://localhost:26657",
		path.Join(usr.HomeDir, ".mpcli"))
	if err != nil {
		t.Errorf("failed to create cliCtx connection: %v", err)
		return
	}
	cliCtx = cliCtx.WithCodec(cdc)

	handler := handlers.NewMarketplaceHandler(cliCtx)
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
