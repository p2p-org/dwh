package indexer

import (
	"context"
	"strings"
	"time"

	cliCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/common/log"
)

type Indexer struct {
	ctx       cliCtx.CLIContext
	txDecoder sdk.TxDecoder
	db        *gorm.DB
	handlers  map[string]MsgHandler
}

func NewIndexer(
	ctx cliCtx.CLIContext,
	txDecoder sdk.TxDecoder,
	db *gorm.DB,
	handlers map[string]MsgHandler,
) *Indexer {
	return &Indexer{
		ctx:       ctx,
		txDecoder: txDecoder,
		db:        db,
		handlers:  handlers,
	}
}

func (m *Indexer) Run(ctx context.Context) error {
	rpcClient, err := m.ctx.GetNode()
	if err != nil {
		log.Fatalf("failed to get rpc client: %v", err)
	}

	var height int64 = 1
	for {
		select {
		case <-ctx.Done():
			log.Info("context cancelled, exiting")
			return nil
		default:
			block, err := rpcClient.Block(&height)
			if err != nil {
				// If we query for a block that does not exist yet, we do not want to
				// log the error.
				// TODO: check if the error is named (strings comparison is fuck ugly).
				if !strings.Contains(err.Error(), "Height") {
					log.Errorf("failed to get block at height %d: %v", height, err)
				}
				time.Sleep(time.Second)
				continue
			}

			log.Infof("retrieved block #%d, block ID %s, transactions: %d",
				height, block.BlockMeta.BlockID, block.BlockMeta.Header.NumTxs)

			for txID, txBytes := range block.Block.Data.Txs {
				log.Infof("processing transaction #%d", txID)

				tx, err := m.txDecoder(txBytes)
				if err != nil {
					log.Errorf("failed to decode transaction bytes: %v", err)
				}

				for _, msg := range tx.GetMsgs() {
					handler, ok := m.handlers[msg.Route()]
					if !ok {
						log.Errorf("unknown message route %s (type %s), skipping", msg.Route(), msg.Type())
						continue
					}

					if err := handler(msg); err != nil {
						log.Errorf("failed to process message %+v", msg)
					}
				}
			}

			height++
		}
	}
}
