package dwh_common

import (
	"fmt"
	stdLog "log"
	"net/url"

	"github.com/spf13/viper"
)

const (
	DefaultConfigName = "config"
	DefaultConfigPath = "/root/"
)

type IndexerCfg struct {
	StatePath string `mapstructure:"state_path"`
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

type DwhCommonServiceConfig struct {
	IndexerCfg              `mapstructure:"indexer"`
	RabbitMQCfg             `mapstructure:"rabbitmq"`
	ImgResizerServiceCfg    `mapstructure:"img_resizer_service"`
	ImgStorageServiceCfg    `mapstructure:"img_storage_service"`
	TokenMetaDataServiceCfg `mapstructure:"token_metadata_service"`
	MongoDaemonServiceCfg   `mapstructure:"mongo_daemon_service"`
	MongoDBCfg              `mapstructure:"mongo_db"`
}

func DefaultDwhCommonServiceConfig() *DwhCommonServiceConfig {
	return &DwhCommonServiceConfig{
		IndexerCfg: IndexerCfg{
			StatePath: "./indexer.state",
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
