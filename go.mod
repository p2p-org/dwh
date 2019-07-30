module github.com/dgamingfoundation/dwh

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/cosmos/sdk-application-tutorial v0.0.0-20190624153636-5544cb2b56cc // indirect
	github.com/dgamingfoundation/dkglib v0.0.0-20190730110748-bffec4019cb8
	github.com/dgamingfoundation/marketplace v0.0.0-20190723150646-fe2fd52c8909
	github.com/google/uuid v1.1.1
	github.com/jinzhu/gorm v1.9.10
	github.com/lib/pq v1.1.1
	github.com/prometheus/common v0.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.7
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.0.0-20190625145210-5fb86c661ea1
