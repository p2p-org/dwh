package auctionHelper_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ah "github.com/dgamingfoundation/dwh/x/auctionHelper"
)

func recordsAreEqual(ar1, ar2 *ah.AuctionLotRecord) bool {
	if ar1 == nil || ar2 == nil {
		return false
	}
	return ar1.ExpirationTime.Equal(ar2.ExpirationTime)
}

func slicesAreEqual(s1, s2 []*ah.AuctionLotRecord) bool {
	if len(s1) != len(s2) {
		return false
	}

	for k := 0; k < len(s1); k++ {
		if !recordsAreEqual(s1[k], s2[k]) {
			return false
		}
	}

	return true
}

func TestAuctionHelper_InsertLot(t *testing.T) {
	ctx := context.Background()
	hlpr := ah.NewAuctionHelper(ctx, nil, nil)
	for _, v := range testSet1 {
		v := v
		hlpr.InsertLot(v.TokenID, v.ExpirationTime)
	}
	var list1 []*ah.AuctionLotRecord
	list1 = append(list1, testSet1...)
	list2 := hlpr.GetExpiredList()

	sort.SliceStable(list1, func(i, j int) bool {
		return list1[i].ExpirationTime.UTC().UnixNano() < list1[j].ExpirationTime.UTC().UnixNano()
	})
	ok := slicesAreEqual(list1, list2)
	assert.True(t, ok)
}

func TestAuctionHelper_RemoveLot(t *testing.T) {
	ctx := context.Background()
	hlpr := ah.NewAuctionHelper(ctx, nil, nil)
	for _, v := range testSet2 {
		v := v
		hlpr.InsertLot(v.TokenID, v.ExpirationTime)
	}
	list1 := hlpr.GetExpiredList()
	for _, v := range list1 {
		v := v
		hlpr.RemoveLot(v.TokenID, v.ExpirationTime)
	}

	list1 = hlpr.GetCurrentList()
	ok := slicesAreEqual(list1, testSet2Check)
	assert.True(t, ok)
	for i := 0; i < len(list1); i++ {
		t.Log(list1[i])
	}
}

var baseTime1 = time.Now().UTC().Add(time.Second * -5)

var testSet1 = []*ah.AuctionLotRecord{
	&ah.AuctionLotRecord{TokenID: "1", ExpirationTime: baseTime1},
	&ah.AuctionLotRecord{TokenID: "2", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "3", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "4", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "5", ExpirationTime: baseTime1.Add(time.Second * 2)},
	&ah.AuctionLotRecord{TokenID: "6", ExpirationTime: baseTime1.Add(time.Second * 2)},
	&ah.AuctionLotRecord{TokenID: "7", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "8", ExpirationTime: baseTime1},
	&ah.AuctionLotRecord{TokenID: "9", ExpirationTime: baseTime1},
	&ah.AuctionLotRecord{TokenID: "10", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "11", ExpirationTime: baseTime1.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "12", ExpirationTime: baseTime1.Add(time.Second * 3)},
	&ah.AuctionLotRecord{TokenID: "13", ExpirationTime: baseTime1.Add(time.Second * 3)},
	&ah.AuctionLotRecord{TokenID: "14", ExpirationTime: baseTime1.Add(time.Second * 3)},
	&ah.AuctionLotRecord{TokenID: "15", ExpirationTime: baseTime1.Add(time.Second * 3)},
	&ah.AuctionLotRecord{TokenID: "16", ExpirationTime: baseTime1.Add(time.Second * 3)},
}

var baseTime2 = time.Now().UTC()

var testSet2 = []*ah.AuctionLotRecord{
	&ah.AuctionLotRecord{TokenID: "1", ExpirationTime: baseTime2.Add(time.Second * -1)},
	&ah.AuctionLotRecord{TokenID: "2", ExpirationTime: baseTime2},
	&ah.AuctionLotRecord{TokenID: "3", ExpirationTime: baseTime2.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "4", ExpirationTime: baseTime2.Add(time.Second * -1)},
	&ah.AuctionLotRecord{TokenID: "5", ExpirationTime: baseTime2},
	&ah.AuctionLotRecord{TokenID: "6", ExpirationTime: baseTime2},
	&ah.AuctionLotRecord{TokenID: "7", ExpirationTime: baseTime2.Add(time.Second * -1)},
	&ah.AuctionLotRecord{TokenID: "8", ExpirationTime: baseTime2.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "9", ExpirationTime: baseTime2.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "10", ExpirationTime: baseTime2.Add(time.Second * -1)},
}

var testSet2Check = []*ah.AuctionLotRecord{
	&ah.AuctionLotRecord{TokenID: "3", ExpirationTime: baseTime2.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "8", ExpirationTime: baseTime2.Add(time.Second * 1)},
	&ah.AuctionLotRecord{TokenID: "9", ExpirationTime: baseTime2.Add(time.Second * 1)},
}
