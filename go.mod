module github.com/dgamingfoundation/dwh

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.0.0-20190625145210-5fb86c661ea1
	github.com/dgamingfoundation/marketplace v0.0.0-20190730173039-2eb669ce5fd8
	github.com/google/uuid v1.1.1
	github.com/jinzhu/gorm v1.9.10
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lib/pq v1.1.1
	github.com/prometheus/common v0.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
