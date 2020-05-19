module github.com/p2p-org/dwh

go 1.12

require (
	github.com/corestario/cosmos-utils/client v0.0.0-20200518174750-e20fa2abb463
	github.com/corestario/marketplace v0.0.0-20191227095102-07bb99b5023b // indirect
	github.com/cosmos/cosmos-sdk v0.38.0
	github.com/cosmos/modules/incubator/nft v0.0.0-20200409061055-9d5a3d97f9b1
	github.com/gorilla/mux v1.7.4
	github.com/h2non/filetype v1.0.10
	github.com/h2non/go-is-svg v0.0.0-20160927212452-35e8c4b0612c
	github.com/jinzhu/gorm v1.9.10
	github.com/lib/pq v1.1.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/p2p-org/marketplace v0.0.0
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/viper v1.6.3
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.4
	github.com/tidwall/pretty v1.0.0 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.1.0
	go.mongodb.org/mongo-driver v1.1.1
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
)

replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.34.4-0.20200507135526-b3cada10017d

replace github.com/p2p-org/marketplace => ./../marketplace
replace github.com/cosmos/modules/incubator/nft => ./../modules/incubator/nft
