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
	mpChain      = "mpchain"
	mpBaseAddr   = "http://localhost:1317/"
	mpFinishAddr = mpBaseAddr + "marketplace/finish_auction"
	//mpAccount    = "cosmos1wfq8q930gh0pktf4h9ddq20czgs8j9yh4tahs0" //user1
	//mpAccount = "cosmos1d7y99yghc2dr56wprqg2zm9j3cnycr22alr8r4" //user2
	mpAccount     = "cosmos1tctr64k4en25uvet2k2tfkwkh0geyrv8fvuvet" //dgaming
	mpAccountAddr = mpBaseAddr + "auth/accounts/" + mpAccount
)

type AuctionLotRecord struct {
	TokenID        string
	ExpirationTime time.Time
}

type AuctionHelper struct {
	mu              sync.RWMutex
	tokenMap        map[string]time.Time           // map for fast-check existing auctions
	lotSlice        []*AuctionLotRecord            // slice for fast data access; sorted by time
	ctx             context.Context                // Global context for Indexer.
	cfg             *common.DwhCommonServiceConfig // Config for all services
	cancel          context.CancelFunc             // Used to stop main processing loop.
	db              *gorm.DB                       // Database to store data to.
	memo            string
	accountSequence uint64
	accountNumber   uint64
	httpClient      *http.Client
}

func NewAuctionHelper(
	ctx context.Context,
	cfg *common.DwhCommonServiceConfig,
	db *gorm.DB,
) (*AuctionHelper, error) {
	ctx, cancel := context.WithCancel(ctx)
	hlpr := &AuctionHelper{
		tokenMap:   make(map[string]time.Time),
		lotSlice:   make([]*AuctionLotRecord, 0),
		ctx:        ctx,
		cfg:        cfg,
		cancel:     cancel,
		db:         db,
		httpClient: &http.Client{Timeout: time.Second * 10},
	}

	err := hlpr.getAccount()
	if err != nil {
		return nil, err
	}

	return hlpr, nil
}

func (ah *AuctionHelper) Run() {
	log.Println("start auction helper main cycle")
	readTicker := time.NewTicker(time.Minute * 1)
	sendTicker := time.NewTicker(time.Second * 5)
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
	log.Println("start get lots from DB")
	var nfts []common.NFT
	ah.db.Where("status = ? AND time_to_sell < ?", 2, time.Now().UTC().Add(time.Minute*5)).Find(&nfts)
	for _, v := range nfts {
		v := v
		ah.insertLot(v.TokenID, v.TimeToSell)
	}
	log.Println("nfts: ", len(nfts))
}

func (ah *AuctionHelper) insertLot(id string, expTime time.Time) bool {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	if _, ok := ah.tokenMap[id]; ok {
		return false
	}
	ah.tokenMap[id] = expTime
	i := sort.Search(len(ah.lotSlice), func(i int) bool { return ah.lotSlice[i].ExpirationTime.Nanosecond() >= expTime.Nanosecond() })
	ah.lotSlice = append(ah.lotSlice, nil)
	copy(ah.lotSlice[i+1:], ah.lotSlice[i:])
	ah.lotSlice[i] = &AuctionLotRecord{TokenID: id, ExpirationTime: expTime}
	return true
}

func (ah *AuctionHelper) removeLot(id string, expTime time.Time) bool {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	i := sort.Search(len(ah.lotSlice), func(i int) bool { return ah.lotSlice[i].ExpirationTime.Nanosecond() >= expTime.Nanosecond() })
	if i < len(ah.lotSlice) && ah.lotSlice[i].TokenID == id {
		if i < len(ah.lotSlice)-1 {
			copy(ah.lotSlice[i:], ah.lotSlice[i+1:])
		}
		ah.lotSlice[len(ah.lotSlice)-1] = nil
		ah.lotSlice = ah.lotSlice[:len(ah.lotSlice)-1]
	} else {
		log.Println("ERROR: element not found!")
		return false
	}
	delete(ah.tokenMap, id)

	return true
}

func (ah *AuctionHelper) FinishAuctions() (out error) {
	log.Println("start finish auction lots")
	list := ah.getExpiredList()
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
		ah.removeLot(v.TokenID, v.ExpirationTime)
	}
	return
}

func (ah *AuctionHelper) getExpiredList() []*AuctionLotRecord {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	t := time.Now().UTC()
	var out []*AuctionLotRecord
	i := sort.Search(len(ah.lotSlice), func(i int) bool { return ah.lotSlice[i].ExpirationTime.Nanosecond() >= t.Nanosecond() })

	out = append([]*AuctionLotRecord{}, ah.lotSlice[:i]...)
	//ah.lotSlice = append([]*AuctionLotRecord{}, ah.lotSlice[i:]...)
	//for _, v := range out {
	//	v := v
	//	delete(ah.tokenMap, v.TokenID)
	//}

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
		Name:     "dgaming",
		Password: "12345678",

		TokenID: id,
	}
	ah.mu.RUnlock()

	ba, err := json.Marshal(&far)
	if err != nil {
		return err
	}
	log.Println(string(ba))

	buf := bytes.NewBuffer(ba)
	req, err := http.NewRequest("PUT", mpFinishAddr, buf)
	if err != nil {
		return err
	}

	resp, err := ah.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(rba))
	if resp.StatusCode != 200 {
		return fmt.Errorf("http error, status code: %v", resp.StatusCode)
	}

	return nil
}

func (ah *AuctionHelper) getAccount() error {
	resp, err := ah.httpClient.Get(mpAccountAddr)
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

type AccountResponse struct {
	Result struct {
		Value struct {
			AccountNumber uint64 `json:"account_number,string"`
			Sequence      uint64 `json:"sequence,string"`
		} `json:"value"`
	} `json:"result"`
}

/*
{
  "height": "181",
  "result": {
    "type": "cosmos-sdk/Account",
    "value": {
      "address": "cosmos1tctr64k4en25uvet2k2tfkwkh0geyrv8fvuvet",
      "coins": [
        {
          "denom": "stake",
          "amount": "100000000"
        },
        {
          "denom": "token",
          "amount": "1000"
        }
      ],
      "public_key": null,
      "account_number": "7",
      "sequence": "0"
    }
  }
}
*/

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
