package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	sdknft "github.com/cosmos/cosmos-sdk/x/nft/types"
	mpnft "github.com/dgamingfoundation/marketplace/x/marketplace/types"
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

func GetDB() (*gorm.DB, error) {
	return gorm.Open("postgres", ConnString)
}

func CreateTables(db *gorm.DB, reset bool) *gorm.DB {
	if reset {
		db = db.DropTableIfExists(&NFT{})
		db = db.DropTableIfExists(&Message{})
		db = db.DropTableIfExists(&User{})

		db = db.CreateTable(&NFT{})
		db = db.CreateTable(&Message{})
		db = db.CreateTable(&User{})
		db = db.AutoMigrate(&NFT{}, &Message{}, &User{})
		db = db.Model(&NFT{}).AddForeignKey(
			"owner_address", "users(address)", "CASCADE", "CASCADE")
	}

	return db
}

func PopulateMockNFTs(numNFTs int64) []*NFT {
	var (
		nfts []*NFT
		idx  int64
	)

	for idx = 0; idx < numNFTs; idx++ {
		_, owner, _, _ := mock.CreateGenAccounts(1, sdk.Coins{
			sdk.NewCoin("mpcoin", sdk.NewInt(idx*100)),
		})
		var (
			name        = fmt.Sprintf("name_%d", idx)
			description = fmt.Sprintf("description_%d", idx)
			image       = fmt.Sprintf("http://image.com/%d", idx)
		)
		nft := &mpnft.NFT{
			BaseNFT: sdknft.NewBaseNFT(
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
		log.Infof("Populating nft:\n%v\n\n", nft.String())
		nfts = append(nfts, NewNFTFromMarketplaceNFT(nft))
	}

	return nfts
}
