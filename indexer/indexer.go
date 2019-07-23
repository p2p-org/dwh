package indexer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	cliCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

const (
	cursorKey = "cursor"
)

var (
	errCursor = errors.New("fatal: failed to update indexer cursor, state is inconsistent")
)

type Indexer struct {
	mu        sync.Mutex
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
		mu:        sync.Mutex{},
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
			return nil, errCursor
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

			if err := m.processTxs(rpcClient, block.Block.Data.Txs); err != nil {
				return fmt.Errorf("failed to processTxs: %v", err)
			}
		}
	}
}

func (m *Indexer) processTxs(rpcClient client.Client, txs types.Txs) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, txBytes := range txs {
		txRes, err := rpcClient.Tx(txBytes.Hash(), true)
		if err != nil {
			log.Debugf("failed to get transaction %s: %v", txBytes.String(), err)
			continue
		}
		if sdk.CodeType(txRes.TxResult.Code) == sdk.CodeUnknownRequest {
			log.Debugf("transaction %s failed (code %d), skipping", txBytes.String(), txRes.TxResult.Code)
			continue
		}
		if txRes.Index < m.cursor.TxIndex {
			log.Debugf("old transaction (%d < %d), skipping", txRes, m.cursor.TxIndex)
			continue
		}
		log.Infof("processing transaction #%d at height %d", txRes.Index, txRes.Height)

		var dbTx = common.NewTx(txRes)
		m.db = m.db.Create(dbTx).Scan(dbTx)
		if m.db.Error != nil {
			log.Errorf("failed to store transaction: %v", err)
			continue
		}
		tx, err := m.txDecoder(txBytes)
		if err != nil {
			log.Errorf("failed to decode transaction bytes: %v", err)
			continue
		}

		for msgID, msg := range tx.GetMsgs() {
			if err := m.processMsg(dbTx.ID, dbTx.Index, msgID, msg); err != nil {
				if err == errCursor {
					// This is a fatal error, indexer should be stopped.
					return err
				}
				log.Errorf("failed to process message: %v", err)
			}
		}
	}
	if err := m.updateCursor(m.cursor.Height+1, 0, 0); err != nil {
		// This is a fatal error, indexer should be stopped.
		return errCursor
	}

	return nil
}

func (m *Indexer) processMsg(txID uint, txIndex uint32, msgID int, msg sdk.Msg) error {
	if msgID < m.cursor.MsgID {
		log.Debugf("old message (%d < %d), skipping", msgID, m.cursor.MsgID)
		return nil
	}

	// We store general information about a message regardless of whether we processed it
	// successfully or not; in case of failure we store additional information about the
	// error.
	var (
		failed bool
		errMsg string
	)
	defer func() {
		m.db = m.db.Create(
			common.NewMessage(
				msg.Route(),
				msg.Type(),
				fmt.Sprintf("%s", msg.GetSignBytes()),
				msg.GetSigners(),
				failed,
				errMsg,
				txID,
			),
		)
		if m.db.Error != nil {
			log.Errorf("failed to add auto migrate: %v", m.db.Error)
		}
	}()

	handler, ok := m.handlers[msg.Route()]
	if !ok {
		failed, errMsg = true, fmt.Sprintf(
			"unknown message route %s (type %s), skipping", msg.Route(), msg.Type())
		return errors.New(errMsg)
	}

	if err := handler.Handle(msg); err != nil {
		failed, errMsg = true, fmt.Sprintf("failed to process message %+v: %v", msg, err)
		return errors.New(errMsg)
	}

	if err := m.updateCursor(m.cursor.Height, txIndex, msgID); err != nil {
		return errCursor
	}

	return nil
}

func (m *Indexer) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.db.Close(); err != nil {
		log.Errorf("failed to close database connection: %v", err)
	}
	if err := m.stateDB.Close(); err != nil {
		log.Errorf("failed to close state database connection: %v", err)
	}
	m.cancel()
}

func (m *Indexer) updateCursor(height int64, txIndex uint32, msgID int) error {
	m.cursor.Height, m.cursor.TxIndex, m.cursor.MsgID = height, txIndex, msgID
	return m.stateDB.Put([]byte(cursorKey), m.cursor.Marshal(), nil)
}
