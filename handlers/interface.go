package handlers

import (
	sdk "github.com/dgamingfoundation/cosmos-sdk/types"
	"github.com/jinzhu/gorm"
)

// MsgHandler is an interface for a handler used by Indexer to process messages
// that belong to various modules. Modules are distinguished by their RouterKey
// (e.g., cosmos-sdk/x/auth.RouterKey).
//
// A handler is supposed to process values of type sdk.Msg using the DB
// connection that is utilized by Indexer.
type MsgHandler interface {
	Handle(db *gorm.DB, msg sdk.Msg) error
	// Setup is meant to prepare the storage. For example, you can create necessary tables
	// and indices for your module here.
	Setup(db *gorm.DB) (*gorm.DB, error)
	// Reset is meant to clear the storage. For example, it is supposed to drop any tables
	// and indices created by the handler.
	Reset(db *gorm.DB) (*gorm.DB, error)
	// RouterKey should return the RouterKey that is used in messages for handler's
	// module.
	// Note: the reason why we use RouterKey (not ModuleName) is because CosmosSDK
	// does not force developers to use ModuleName as RouterKey for registered
	// messages (even though most modules do so).
	RouterKey() string
}
