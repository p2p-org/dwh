module github.com/dgamingfoundation/dwh

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/dgamingfoundation/dkglib v0.0.0-20190729123057-af9cb9815663
	github.com/dgamingfoundation/marketplace v0.0.0-20190723150646-fe2fd52c8909
	github.com/golang/mock v1.3.1-0.20190508161146-9fa652df1129 // indirect
	github.com/google/uuid v1.1.1
	github.com/graphql-go/graphql v0.7.8
	github.com/jinzhu/gorm v1.9.10
	github.com/lib/pq v1.1.1
	github.com/prometheus/common v0.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.7
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.28.2-0.20190616100639-18415eedaf25
