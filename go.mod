module github.com/dgamingfoundation/dwh

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.0.0-20190625145210-5fb86c661ea1
	github.com/dgamingfoundation/marketplace v0.0.0-20190716124432-f7e460c5ef8e
	github.com/google/uuid v1.1.1
	github.com/graphql-go/graphql v0.7.8 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/lib/pq v1.1.1
	github.com/prometheus/common v0.2.0
	github.com/sirupsen/logrus v1.4.2
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
