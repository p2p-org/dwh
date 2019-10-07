package mongoDaemon

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

type DwhQueueServiceConfig struct {
	QueueScheme   string `mapstructure:"queue_scheme"`
	QueueAddr     string `mapstructure:"queue_addr"`
	QueuePort     int    `mapstructure:"queue_port"`
	QueuePath     string `mapstructure:"queue_path"`
	QueueUsername string `mapstructure:"queue_username"`
	QueuePassword string `mapstructure:"queue_password"`

	ExchangeName string `mapstructure:"exchange_name"`

	UriQueueName          string `mapstructure:"uri_queue_name"`
	UriQueueMaxPriority   int    `mapstructure:"uri_max_priority"`
	UriQueuePrefetchCount int    `mapstructure:"uri_prefetch_count"`

	DaemonTaskQueueName          string `mapstructure:"daemon_task_queue_name"`
	DaemonDelayedQueueName       string `mapstructure:"daemon_delayed_task_queue_name"`
	DaemonTaskQueueMaxPriority   int    `mapstructure:"daemon_task_max_priority"`
	DaemonTaskQueuePrefetchCount int    `mapstructure:"daemon_task_prefetch_count"`
	DaemonTTLSeconds             int    `mapstructure:"daemon_ttl_seconds"`
	DaemonUpdatePercent          int64  `mapstructure:"daemon_update_percent"`

	MongoUserName   string `mapstructure:"mongo_user_name"`
	MongoUserPass   string `mapstructure:"mongo_user_pass"`
	MongoHost       string `mapstructure:"mongo_host"`
	MongoDatabase   string `mapstructure:"mongo_database"`
	MongoCollection string `mapstructure:"mongo_collection"`
}

func DefaultDwhQueueServiceConfig() *DwhQueueServiceConfig {
	return &DwhQueueServiceConfig{
		QueueScheme:   "amqp",
		QueueAddr:     "localhost",
		QueuePort:     5672,
		QueueUsername: "guest",
		QueuePassword: "guest",

		ExchangeName: "dwh_direct_exchange",

		UriQueueName:          "dwh_uri_tasks",
		UriQueueMaxPriority:   10,
		UriQueuePrefetchCount: 1,

		DaemonTaskQueueName:          "daemon_mongo_tasks",
		DaemonDelayedQueueName:       "daemon_delayed_mongo_tasks",
		DaemonTaskQueueMaxPriority:   10,
		DaemonTaskQueuePrefetchCount: 1,
		DaemonTTLSeconds:             60 * 1,
		DaemonUpdatePercent:          20,

		MongoUserName:   "dgaming",
		MongoUserPass:   "dgaming",
		MongoHost:       "localhost:27017",
		MongoDatabase:   "dgaming",
		MongoCollection: "token_metadata",
	}
}

func QueueAddrStringFromConfig(cfg *DwhQueueServiceConfig) string {
	u := url.URL{
		Scheme: cfg.QueueScheme,
		User:   url.UserPassword(cfg.QueueUsername, cfg.QueuePassword),
		Host:   fmt.Sprintf("%s:%d", cfg.QueueAddr, cfg.QueuePort),
		Path:   cfg.QueuePath,
	}
	return u.String()
}

func ReadDwhQueueServiceConfig(configName, path string) *DwhQueueServiceConfig {
	var cfg *DwhQueueServiceConfig
	vCfg := viper.New()
	vCfg.SetConfigName(configName)
	vCfg.AddConfigPath(path)
	err := vCfg.ReadInConfig()
	if err != nil {
		fmt.Println("ERROR: server config file not found, error:", err)
		return DefaultDwhQueueServiceConfig()
	}
	err = vCfg.Unmarshal(&cfg)
	if err != nil {
		fmt.Println("ERROR: could not unmarshal server config file, error:", err)
		return DefaultDwhQueueServiceConfig()
	}

	return cfg
}
