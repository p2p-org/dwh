package dwh_common

import (
	"fmt"
	stdLog "log"
	"net/url"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const (
	IndexerConfigFileName = "indexer"
)

const (
	PrometheusEnabledFlag  = "prometheus_enabled"
	PrometheusHostPortFlag = "prort=5432 user=dgaming password=dgaming dbname=marketplace sslmode=disableometheus_host_port"
	PprofEnabledFlag       = "pprof_enabled"
	PprofHostPortFlag      = "pprof_host_port"
	VfrHomeFlag            = "vfr_home"
	HeightFlag             = "height"
	TrustNodeFlag          = "trust_node"
	BroadcastModeFlag      = "broadcast_mode"
	GenOnlyFlag            = "gen_only"
	UserNameFlag           = "user_name"
)

const (
	DefaultConfigName = "config"
	DefaultConfigPath = "/root/"
)

type IndexerCfg struct {
	StatePath       string `mapstructure:"state_path"`
	ResetDatabase   bool   `mapstructure:"reset_database"`
	MarketplaceAddr string `mapstructure:"marketplace_addr"`
	ChainID         string `mapstructure:"chain_id"`
	CliHome         string `mapstructure:"cli_home"`
}

type RabbitMQCfg struct {
	QueueScheme   string `mapstructure:"queue_scheme"`
	QueueAddr     string `mapstructure:"queue_addr"`
	QueuePort     int    `mapstructure:"queue_port"`
	QueuePath     string `mapstructure:"queue_path"`
	QueueUsername string `mapstructure:"queue_username"`
	QueuePassword string `mapstructure:"queue_password"`
	ExchangeName  string `mapstructure:"exchange_name"`
}

type ImgResizerServiceCfg struct {
	ImgQueueName          string       `mapstructure:"img_queue_name"`
	ImgQueueMaxPriority   int          `mapstructure:"img_max_priority"`
	ImgQueuePrefetchCount int          `mapstructure:"img_prefetch_count"`
	InterpolationMethod   int          `mapstructure:"interpolation_method"` // 2
	Resolutions           []Resolution `mapstructure:"resolutions"`
}

type ImgStorageServiceCfg struct {
	StorageAddr             string `mapstructure:"storage_addr"`
	StoragePort             int    `mapstructure:"storage_port"`
	StorageCompressedOption bool   `mapstructure:"storage_is_compressed"`
	StorageDiskPath         string `mapstructure:"storage_path"`
}

type TokenMetaDataServiceCfg struct {
	UriQueueName          string `mapstructure:"uri_queue_name"`
	UriQueueMaxPriority   int    `mapstructure:"uri_max_priority"`
	UriQueuePrefetchCount int    `mapstructure:"uri_prefetch_count"`
}

type MongoDaemonServiceCfg struct {
	DaemonTaskQueueName          string `mapstructure:"daemon_task_queue_name"`
	DaemonDelayedQueueName       string `mapstructure:"daemon_delayed_task_queue_name"`
	DaemonTaskQueueMaxPriority   int    `mapstructure:"daemon_task_max_priority"`
	DaemonTaskQueuePrefetchCount int    `mapstructure:"daemon_task_prefetch_count"`
	DaemonTTLSeconds             int    `mapstructure:"daemon_ttl_seconds"`
	DaemonUpdatePercent          int64  `mapstructure:"daemon_update_percent"`
}

type MongoDBCfg struct {
	MongoUserName   string `mapstructure:"mongo_user_name"`
	MongoUserPass   string `mapstructure:"mongo_user_pass"`
	MongoHost       string `mapstructure:"mongo_host"`
	MongoDatabase   string `mapstructure:"mongo_database"`
	MongoCollection string `mapstructure:"mongo_collection"`
}

type PostgresCfg struct {
	PostgresUserName string `mapstructure:"postgres_user_name"`
	PostgresUserPass string `mapstructure:"postgres_user_pass"`
	PostgresHost     string `mapstructure:"postgres_host"`
	PostgresPort     int    `mapstructure:"postgres_port"`
	PostgresDBName   string `mapstructure:"postgres_db_name"`
}

type DwhCommonServiceConfig struct {
	IndexerCfg              `mapstructure:"indexer"`
	RabbitMQCfg             `mapstructure:"rabbitmq"`
	ImgResizerServiceCfg    `mapstructure:"img_resizer_service"`
	ImgStorageServiceCfg    `mapstructure:"img_storage_service"`
	TokenMetaDataServiceCfg `mapstructure:"token_metadata_service"`
	MongoDaemonServiceCfg   `mapstructure:"mongo_daemon_service"`
	MongoDBCfg              `mapstructure:"mongo_db"`
	PostgresCfg             `mapstructure:"postgres_db"`
}

func DefaultDwhCommonServiceConfig() *DwhCommonServiceConfig {
	return &DwhCommonServiceConfig{
		IndexerCfg: IndexerCfg{
			StatePath:       "./indexer.state",
			ResetDatabase:   false,
			MarketplaceAddr: "tcp://localhost:26657",
			CliHome:         ".mpcli",
			ChainID:         "mpchain",
		},

		RabbitMQCfg: RabbitMQCfg{
			QueueScheme:   "amqp",
			QueueAddr:     "localhost",
			QueuePort:     5672,
			QueueUsername: "guest",
			QueuePassword: "guest",
			ExchangeName:  "dwh_direct_exchange",
		},

		ImgResizerServiceCfg: ImgResizerServiceCfg{
			ImgQueueName:          "dwh_img_tasks",
			ImgQueueMaxPriority:   10,
			ImgQueuePrefetchCount: 1,
			Resolutions: []Resolution{
				{200, 150},
				{120, 90},
			},
			InterpolationMethod: 2,
		},

		ImgStorageServiceCfg: ImgStorageServiceCfg{
			StorageAddr:             "http://127.0.0.1",
			StoragePort:             11535,
			StorageCompressedOption: false,
			StorageDiskPath:         "/root/dwh_storage",
		},

		TokenMetaDataServiceCfg: TokenMetaDataServiceCfg{
			UriQueueName:          "dwh_uri_tasks",
			UriQueueMaxPriority:   10,
			UriQueuePrefetchCount: 1,
		},

		MongoDaemonServiceCfg: MongoDaemonServiceCfg{
			DaemonTaskQueueName:          "daemon_mongo_tasks",
			DaemonDelayedQueueName:       "daemon_delayed_mongo_tasks",
			DaemonTaskQueueMaxPriority:   10,
			DaemonTaskQueuePrefetchCount: 1,
			DaemonTTLSeconds:             60 * 60 * 6,
			DaemonUpdatePercent:          20,
		},

		MongoDBCfg: MongoDBCfg{
			MongoUserName:   "dgaming",
			MongoUserPass:   "dgaming",
			MongoHost:       "localhost:27017",
			MongoDatabase:   "dgaming",
			MongoCollection: "token_metadata",
		},

		PostgresCfg: PostgresCfg{
			PostgresUserName: "dgaming",
			PostgresUserPass: "dgaming",
			PostgresHost:     "localhost",
			PostgresPort:     5432,
			PostgresDBName:   "marketplace",
		},
	}
}

func QueueAddrStringFromConfig(cfg *DwhCommonServiceConfig) string {
	u := url.URL{
		Scheme: cfg.RabbitMQCfg.QueueScheme,
		User:   url.UserPassword(cfg.RabbitMQCfg.QueueUsername, cfg.RabbitMQCfg.QueuePassword),
		Host:   fmt.Sprintf("%s:%d", cfg.RabbitMQCfg.QueueAddr, cfg.RabbitMQCfg.QueuePort),
		Path:   cfg.RabbitMQCfg.QueuePath,
	}
	return u.String()
}

func ReadCommonConfig(configName, path string) *DwhCommonServiceConfig {
	cfg := DefaultDwhCommonServiceConfig()
	vCfg := viper.New()
	vCfg.SetConfigName(configName)
	vCfg.AddConfigPath(path)
	err := vCfg.ReadInConfig()
	if err != nil {
		stdLog.Println("server config file not found, load default config")
		return cfg
	}
	err = vCfg.Unmarshal(&cfg)
	if err != nil {
		stdLog.Println("could not unmarshal server config file, load default config")
		return cfg
	}

	return cfg
}

func InitConfig() {
	viper.SetDefault(PrometheusEnabledFlag, true)
	viper.SetDefault(PrometheusHostPortFlag, "localhost:9081")
	viper.SetDefault(PprofEnabledFlag, true)
	viper.SetDefault(PprofHostPortFlag, "localhost:6061")
	viper.SetDefault(VfrHomeFlag, "")
	viper.SetDefault(HeightFlag, 0)
	viper.SetDefault(TrustNodeFlag, false)
	viper.SetDefault(BroadcastModeFlag, "sync")
	viper.SetDefault(GenOnlyFlag, false)
	viper.SetDefault(UserNameFlag, "user1")
	viper.SetConfigName(IndexerConfigFileName)
	viper.AddConfigPath("$HOME/.dwh/config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			stdLog.Println("config file not found, using default configuration")
		} else {
			stdLog.Fatalf("failed to parse config file, exiting: %v", err)
		}
	}
}

func GetDB(cfg *DwhCommonServiceConfig) (*gorm.DB, error) {
	ConnString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUserName,
		cfg.PostgresUserPass,
		cfg.PostgresDBName,
	)
	stdLog.Println("ConnString:", ConnString)

	return gorm.Open("postgres", ConnString)
}
