package indexer

import (
	"context"
	"fmt"
	"strings"
	"time"

	cliCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/types"
)

const (
	cursorKey = "cursor"
)

type Indexer struct {
	ctx       context.Context    // Global context for Indexer.
	cancel    context.CancelFunc // Used to stop main processing loop.
	cliCtx    cliCtx.CLIContext  // Cosmos CLIContext, used to talk to node.
	txDecoder sdk.TxDecoder
	db        *gorm.DB              // Database to store data to.
	stateDB   *leveldb.DB           // State database to keep indexer state.
	handlers  map[string]MsgHandler // A map from module name to its handler (e.g., bank, ibc, marketplace, etc.)
	cursor    *cursor               // Indexer cursor (keeps track of the last processed message).
}

func NewIndexer(
	ctx context.Context,
	cfg *Config,
	cliCtx cliCtx.CLIContext,
	txDecoder sdk.TxDecoder,
	db *gorm.DB,
	handlers map[string]MsgHandler,
) (*Indexer, error) {
	ctx, cancel := context.WithCancel(ctx)
	stateDB, err := leveldb.OpenFile(cfg.StatePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open state database: %v", err)
	}

	idxr := &Indexer{
		ctx:       ctx,
		cancel:    cancel,
		cliCtx:    cliCtx,
		txDecoder: txDecoder,
		db:        db,
		stateDB:   stateDB,
		handlers:  handlers,
		cursor:    &cursor{},
	}

	cursorExists, err := stateDB.Has([]byte(cursorKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check for indexer cursor: %v", err)
	}
	if cursorExists {
		bz, err := stateDB.Get([]byte(cursorKey), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve indexer cursor: %v", err)
		}
		if err := idxr.cursor.Unmarshal(bz); err != nil {
			return nil, fmt.Errorf("failed to unmarshal indexer cursor: %v", err)
		}
	} else {
		if err := idxr.updateCursor(1, 0, 0); err != nil {
			return nil, fmt.Errorf("failed to update indexer cursor, state is incosistent: %v", err)
		}
	}

	return idxr, nil
}

func (m *Indexer) Start() error {
	rpcClient, err := m.cliCtx.GetNode()
	if err != nil {
		log.Fatalf("failed to get rpc client: %v", err)
	}

	for {
		select {
		case <-m.ctx.Done():
			log.Info("context cancelled, exiting")
			return nil
		default:
			block, err := rpcClient.Block(&m.cursor.Height)
			if err != nil {
				// If we query for a block that does not exist yet, we do not want to
				// log the error.
				// TODO: check if the error is named (strings comparison is fuck ugly).
				if !strings.Contains(err.Error(), "Height") {
					log.Errorf("failed to get block at height %d: %v", m.cursor.Height, err)
				}
				time.Sleep(time.Second)
				continue
			}

			log.Infof("retrieved block #%d, block ID %s, transactions: %d",
				m.cursor.Height, block.BlockMeta.BlockID, block.BlockMeta.Header.NumTxs)

			if err := m.processTxs(block.Block.Data.Txs); err != nil {
				return fmt.Errorf("failed to processTxs: %v", err)
			}
		}
	}
}

func (m *Indexer) processTxs(txs types.Txs) error {
	for txID, txBytes := range txs {
		if txID < m.cursor.TxID {
			log.Debugf("old transaction (%d < %d), skipping", txID, m.cursor.TxID)
			continue
		}
		log.Infof("processing transaction #%d", txID)

		tx, err := m.txDecoder(txBytes)
		if err != nil {
			log.Errorf("failed to decode transaction bytes: %v", err)
			continue
		}

		for msgID, msg := range tx.GetMsgs() {
			if msgID < m.cursor.MsgID {
				log.Debugf("old message (%d < %d), skipping", msgID, m.cursor.MsgID)
				continue
			}
			handler, ok := m.handlers[msg.Route()]
			if !ok {
				log.Errorf("unknown message route %s (type %s), skipping", msg.Route(), msg.Type())
				continue
			}

			if err := handler(msg); err != nil {
				log.Errorf("failed to process message %+v", msg)
			}

			if err := m.updateCursor(m.cursor.Height, txID, msgID); err != nil {
				return fmt.Errorf("failed to update indexer cursor, state is incosistent: %v", err)
			}
		}
	}
	if err := m.updateCursor(m.cursor.Height+1, 0, 0); err != nil {
		return fmt.Errorf("failed to update indexer cursor, state is incosistent: %v", err)
	}

	return nil
}

func (m *Indexer) Stop() {
	if err := m.db.Close(); err != nil {
		log.Errorf("failed to close database connection: %v", err)
	}
	if err := m.stateDB.Close(); err != nil {
		log.Errorf("failed to close state database connection: %v", err)
	}
}

func (m *Indexer) updateCursor(height int64, txID, msgID int) error {
	m.cursor.Height, m.cursor.TxID, m.cursor.MsgID = height, txID, msgID
	return m.stateDB.Put([]byte(cursorKey), m.cursor.Marshal(), nil)
}
