package auctionHelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	log "github.com/sirupsen/logrus"

	common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/jinzhu/gorm"
)

const (
	mpChain   = "mpchain"
	mpAccount = "cosmos1tctr64k4en25uvet2k2tfkwkh0geyrv8fvuvet" //dgaming
)

type AuctionLotRecord struct {
	TokenID        string
	ExpirationTime time.Time
}

type AccountResponse struct {
	Result struct {
		Value struct {
			AccountNumber uint64 `json:"account_number,string"`
			Sequence      uint64 `json:"sequence,string"`
		} `json:"value"`
	} `json:"result"`
}

type BaseReq struct {
	From          string       `json:"from,omitempty"`
	Memo          string       `json:"memo,omitempty"`
	ChainID       string       `json:"chain_id,omitempty"`
	AccountNumber uint64       `json:"account_number,string,omitempty"`
	Sequence      uint64       `json:"sequence,string,omitempty"`
	Fees          sdk.Coins    `json:"fees,omitempty"`
	GasPrices     sdk.DecCoins `json:"gas_prices,omitempty"`
	Gas           string       `json:"gas,omitempty"`
	GasAdjustment string       `json:"gas_adjustment,omitempty"`
	Simulate      bool         `json:"simulate,omitempty"`
}

type FinishAuctionReq struct {
	BaseReq BaseReq `json:"base_req"`

	Name     string `json:"name"`
	Password string `json:"password"`

	TokenID string `json:"token_id"`
}

type AuctionHelper struct {
	mu                      sync.RWMutex
	tokenMap                map[string]time.Time           // map for fast-check existing auctions
	lotSlice                []*AuctionLotRecord            // slice for fast data access; sorted by time
	ctx                     context.Context                // Global context for Indexer.
	cfg                     *common.DwhCommonServiceConfig // Config for all services
	cancel                  context.CancelFunc             // Used to stop main processing loop.
	db                      *gorm.DB                       // Database to store data to.
	accountSequence         uint64
	accountNumber           uint64
	httpClient              *http.Client
	finishAddr, accountAddr string
}

func NewAuctionHelper(
	ctx context.Context,
	cfg *common.DwhCommonServiceConfig,
	db *gorm.DB,
) *AuctionHelper {
	baseAddr := fmt.Sprintf("%v:%d/", cfg.AuctionHelperCfg.AuctionHelperMarketplaceHost, cfg.AuctionHelperCfg.AuctionHelperMarketplacePort)

	ctx, cancel := context.WithCancel(ctx)
	hlpr := &AuctionHelper{
		tokenMap:    make(map[string]time.Time),
		lotSlice:    make([]*AuctionLotRecord, 0),
		ctx:         ctx,
		cfg:         cfg,
		cancel:      cancel,
		db:          db,
		httpClient:  &http.Client{Timeout: time.Second * 10},
		finishAddr:  baseAddr + "marketplace/finish_auction",
		accountAddr: baseAddr + "auth/accounts/" + mpAccount,
	}

	return hlpr
}

func (ah *AuctionHelper) Run() {
	err := ah.getAccount()
	if err != nil {
		panic(err)
	}
	log.Println("start auction helper main cycle")
	readTicker := time.NewTicker(time.Second * time.Duration(ah.cfg.AuctionHelperDWHUpdateTickerSeconds))
	sendTicker := time.NewTicker(time.Second * time.Duration(ah.cfg.AuctionHelperFinishTickerSeconds))
	ah.GetLotsFromDB()

	for {
		select {
		case <-readTicker.C:
			ah.GetLotsFromDB()
		case <-sendTicker.C:
			err := ah.FinishAuctions()
			if err != nil {
				log.Println("send error:", err)
			}
			err = ah.getAccount()
			if err != nil {
				panic(err)
			}
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (ah *AuctionHelper) GetLotsFromDB() {
	var nfts []common.NFT
	ah.db.Where("status = ? AND time_to_sell < ?", 2, time.Now().UTC().Add(time.Minute*5)).Find(&nfts)
	for _, v := range nfts {
		v := v
		ah.InsertLot(v.TokenID, v.TimeToSell)
	}
}

func (ah *AuctionHelper) InsertLot(id string, expTime time.Time) bool {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	if _, ok := ah.tokenMap[id]; ok {
		return false
	}
	ah.tokenMap[id] = expTime
	i := sort.Search(len(ah.lotSlice), func(i int) bool { return ah.lotSlice[i].ExpirationTime.UTC().UnixNano() >= expTime.UTC().UnixNano() })
	ah.lotSlice = append(ah.lotSlice, nil)
	copy(ah.lotSlice[i+1:], ah.lotSlice[i:])

	ah.lotSlice[i] = &AuctionLotRecord{
		TokenID:        id,
		ExpirationTime: expTime,
	}
	return true
}

func (ah *AuctionHelper) RemoveLot(id string, expTime time.Time) bool {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	i := sort.Search(len(ah.lotSlice), func(i int) bool { return ah.lotSlice[i].ExpirationTime.UTC().UnixNano() >= expTime.UTC().UnixNano() })
	if i < len(ah.lotSlice) {
		for j := 0; j <= i; j++ {
			if ah.lotSlice[j].TokenID == id {
				if i < len(ah.lotSlice)-1 {
					copy(ah.lotSlice[j:], ah.lotSlice[j+1:])
				}
				ah.lotSlice[len(ah.lotSlice)-1] = nil
				ah.lotSlice = ah.lotSlice[:len(ah.lotSlice)-1]
				delete(ah.tokenMap, id)
				return true
			}
		}
	}

	return false
}

func (ah *AuctionHelper) FinishAuctions() (out error) {
	list := ah.GetExpiredList()
	for _, v := range list {
		v := v
		if err := ah.SendFinish(v.TokenID); err != nil {
			log.Printf("send finish_auction error: %v", err)
			out = err
			continue
		}
		ah.mu.Lock()
		ah.accountSequence++
		ah.mu.Unlock()
		ah.RemoveLot(v.TokenID, v.ExpirationTime)
	}
	return
}

func (ah *AuctionHelper) GetCurrentList() []*AuctionLotRecord {
	ah.mu.RLock()
	defer ah.mu.RUnlock()
	res := append([]*AuctionLotRecord{}, ah.lotSlice...)
	return res
}

func (ah *AuctionHelper) GetExpiredList() []*AuctionLotRecord {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	t := time.Now().UTC()
	var out []*AuctionLotRecord
	i := sort.Search(len(ah.lotSlice), func(i int) bool {
		res := ah.lotSlice[i].ExpirationTime.UTC().UnixNano() >= t.UTC().UnixNano()
		return res
	})

	out = append([]*AuctionLotRecord{}, ah.lotSlice[:i]...)

	return out
}

func (ah *AuctionHelper) SendFinish(id string) error {
	ah.mu.RLock()
	far := FinishAuctionReq{
		BaseReq: BaseReq{
			Sequence:      ah.accountSequence,
			ChainID:       mpChain,
			AccountNumber: ah.accountNumber,
			From:          mpAccount,
			//Gas:           "200000",
		},
		Name:     ah.cfg.AuctionHelperCfg.AuctionHelperName,
		Password: ah.cfg.AuctionHelperCfg.AuctionHelperPassword,

		TokenID: id,
	}
	ah.mu.RUnlock()

	ba, err := json.Marshal(&far)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(ba)
	req, err := http.NewRequest("PUT", ah.finishAddr, buf)
	if err != nil {
		return err
	}

	resp, err := ah.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		rsp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("http error, status code: %v, %v", resp.StatusCode, string(rsp))
	}

	return nil
}

func (ah *AuctionHelper) getAccount() error {
	resp, err := ah.httpClient.Get(ah.accountAddr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var accResp AccountResponse

	if err := json.Unmarshal(ba, &accResp); err != nil {
		return err
	}

	ah.mu.Lock()
	defer ah.mu.Unlock()
	ah.accountNumber = accResp.Result.Value.AccountNumber
	ah.accountSequence = accResp.Result.Value.Sequence
	return nil
}
