package handlers

import (
	sdk "github.com/dgamingfoundation/cosmos-sdk/types"
	"github.com/jinzhu/gorm"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

// MsgHandler is an interface for a handler used by Indexer to process messages
// that belong to various modules. Modules are distinguished by their RouterKey
// (e.g., cosmos-sdk/x/auth.RouterKey).
//
// A handler is supposed to process values of type sdk.Msg using the DB
// connection that is utilized by Indexer.
type MsgHandler interface {
	// Handle is supposed to handle a message along with its associated events.
	// NOTE:  only events that have the same type as the message
	// can be associated with that message.
	Handle(db *gorm.DB, msg sdk.Msg, events ...*abciTypes.Event) error
	// Setup is supposed to prepare the storage. For example, you can create necessary tables
	// and indices for your module here.
	Setup(db *gorm.DB) (*gorm.DB, error)
	// Reset is meant to clear the storage. For example, it is supposed to drop any tables
	// and indices created by the handler.
	Reset(db *gorm.DB) (*gorm.DB, error)
	// RouterKey should return the RouterKeys that are used in messages for handler's
	// module. Multiple keys allow for using the same handler for multiple routes
	// (which might be required if the application intercepts some other module's
	// messages).
	//
	// Note: the reason why we use RouterKey (not ModuleName) is because CosmosSDK
	// does not force developers to use ModuleName as RouterKey for registered
	// messages (even though most modules do so).
	RouterKeys() []string
}
