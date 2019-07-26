package handlers

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgHandler is an interface for a handler used by Indexer to process messages
// that belong to various modules. Modules are distinguished by their RouterKey
// (e.g., cosmos-sdk/x/auth.RouterKey).
// Note: the reason why we use RouterKey (not ModuleName) is because CosmosSDK
// does not force developers to use ModuleName as RouterKey for registered
// messages (even though most modules do so).
//
// A handler is supposed to process values of type sdk.Msg using the DB
// connection that is utilized by Indexer (so that the GraphQL interface could be
// used out of the box), but this is enforced in any way.
type MsgHandler interface {
	Handle(msg sdk.Msg) error
	// Setup is meant to prepare the storage. For example, you can create necessary tables
	// and indices for your module here.
	Setup() error
	RouterKey() string
}
