package main

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	nfttypes "github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/dgamingfoundation/dwh/types"
	mptypes "github.com/dgamingfoundation/marketplace/x/marketplace/types"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var (
	UserName   = "dgaming"
	Password   = "dgaming"
	DBName     = "marketplace"
	ConnString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", UserName, Password, DBName)
)

func main() {
	log.SetLevel(log.DebugLevel)

	db, err := gorm.Open("postgres", ConnString)
	if err != nil {
		log.Fatalf("failed to establish database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
		}
	}()

	db = createTables(db)
	populateMockNFTs(db)
}

func createTables(db *gorm.DB) *gorm.DB {
	db = db.DropTableIfExists()
	return db.CreateTable(&types.NFT{})
}

func populateMockNFTs(db *gorm.DB) {
	var (
		nfts []*types.NFT
		idx  int64
	)

	for idx = 0; idx < 15; idx++ {
		_, owner, _, _ := mock.CreateGenAccounts(1, sdk.Coins{
			sdk.NewCoin("mpcoin", sdk.NewInt(idx*100)),
		})
		var (
			name        = fmt.Sprintf("name_%d", idx)
			description = fmt.Sprintf("description_%d", idx)
			image       = fmt.Sprintf("http://image.com/%d", idx)
		)
		nft := &types.NFT{
			NFT: &mptypes.NFT{
				BaseNFT: nfttypes.NewBaseNFT(
					uuid.New().String(),
					owner[0],
					name,
					description,
					image,
					fmt.Sprintf(`{
    "title": "Asset Metadata",
    "type": "object",
    "properties": {
        "name": {
            "type": "string",
            "description": %s,
        },
        "description": {
            "type": "string",
            "description": %s,
        },
        "image": {
            "type": "string",
            "description": %s,
        }
    }
}`, name, description, image),
				),
			},
		}
		if idx%3 == 0 {
			// Each third NFT is on sale.
			nft.OnSale = true
			nft.Price = sdk.Coins{
				sdk.NewCoin("mpcoin", sdk.NewInt(idx*10)),
			}
			_, sellerBeneficiary, _, _ := mock.CreateGenAccounts(1, sdk.Coins{
				sdk.NewCoin("mpcoin", sdk.NewInt(idx*100)),
			})
			nft.SellerBeneficiary = sellerBeneficiary[0]
		}
		log.Infof("Populating nft:=========\n%v\n", nft.String())
		nfts = append(nfts, nft)
	}

	for _, nft := range nfts {
		db = db.Create(nft)
	}
}
