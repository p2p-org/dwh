package common

import (
	"log"
	"os/user"
	"path"

	"github.com/spf13/viper"
)

const (
	configFileName = "indexer"
)

const (
	PrometheusEnabledFlag  = "prometheus_enabled"
	PrometheusHostPortFlag = "prometheus_host_port"
	PprofEnabledFlag       = "pprof_enabled"
	PprofHostPortFlag      = "pprof_host_port"
	StatePathFlag          = "state_path"
	NodeEndpointFlag       = "node_endpoint"
	ChainIDFlag            = "chain_id"
	VfrHomeFlag            = "vfr_home"
	HeightFlag             = "height"
	TrustNodeFlag          = "trust_node"
	BroadcastModeFlag      = "broadcast_mode"
	GenOnlyFlag            = "gen_only"
	UserNameFlag           = "user_name"
	CliHomeFlag            = "cli_home"
)

func InitConfig() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user, exiting: %v", err)
	}

	viper.SetDefault(PrometheusEnabledFlag, true)
	viper.SetDefault(PrometheusHostPortFlag, "localhost:9081")
	viper.SetDefault(PprofEnabledFlag, true)
	viper.SetDefault(PprofHostPortFlag, "localhost:6061")
	viper.SetDefault(StatePathFlag, path.Join(usr.HomeDir, "indexer.state"))
	viper.SetDefault(NodeEndpointFlag, "tcp://localhost:26657")
	viper.SetDefault(ChainIDFlag, "mpchain")
	viper.SetDefault(VfrHomeFlag, "")
	viper.SetDefault(HeightFlag, 0)
	viper.SetDefault(TrustNodeFlag, false)
	viper.SetDefault(BroadcastModeFlag, "sync")
	viper.SetDefault(GenOnlyFlag, false)
	viper.SetDefault(UserNameFlag, "user1")
	viper.SetDefault(CliHomeFlag, path.Join(usr.HomeDir, ".mpcli"))

	viper.SetConfigName(configFileName)
	viper.AddConfigPath("$HOME/.dwh/config")
	viper.AddConfigPath(".")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("config file not found, using default configuration")
		} else {
			log.Fatalf("failed to parse config file, exiting: %v", err)
		}
	}
}
