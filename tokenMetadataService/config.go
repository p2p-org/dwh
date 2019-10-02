package tokenMetadataService

import (
	"fmt"
	"github.com/dgamingfoundation/dwh/imgservice"
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

	ImgQueuePriority imgservice.ImgQueuePriority `mapstructure:"img_queue_priority"`

	UriQueueName          string `mapstructure:"uri_queue_name"`
	UriQueueMaxPriority   int    `mapstructure:"uri_max_priority"`
	UriQueuePrefetchCount int    `mapstructure:"uri_prefetch_count"`

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

		ImgQueuePriority: 10,

		UriQueueName:          "dwh_uri_tasks",
		UriQueueMaxPriority:   10,
		UriQueuePrefetchCount: 1,

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

func ReadDwhTokenMetadataServiceConfig(configName, path string) *DwhQueueServiceConfig {
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