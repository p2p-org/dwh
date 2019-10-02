package imgservice

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

const DefaultStorePort = 11535

type Resolution struct {
	Width  uint `mapstructure:"width",json:"width"`
	Height uint `mapstructure:"height",json:"height"`
}

type DwhQueueServiceConfig struct {
	QueueScheme   string `mapstructure:"queue_scheme"`
	QueueAddr     string `mapstructure:"queue_addr"`
	QueuePort     int    `mapstructure:"queue_port"`
	QueuePath     string `mapstructure:"queue_path"`
	QueueUsername string `mapstructure:"queue_username"`
	QueuePassword string `mapstructure:"queue_password"`

	ImgQueueName          string `mapstructure:"img_queue_name"`
	ImgQueueMaxPriority   int    `mapstructure:"img_max_priority"`
	ImgQueuePrefetchCount int    `mapstructure:"img_prefetch_count"`

	StoreAddr           string       `mapstructure:"store_addr"`
	StorePort           int          `mapstructure:"store_port"`
	Resolutions         []Resolution `mapstructure:"resolutions"`
	InterpolationMethod int          `mapstructure:"interpolation_method"` // 2
}

func DefaultDwhQueueServiceConfig() *DwhQueueServiceConfig {
	return &DwhQueueServiceConfig{
		QueueScheme:   "amqp",
		QueueAddr:     "localhost",
		QueuePort:     5672,
		QueueUsername: "guest",
		QueuePassword: "guest",

		ImgQueueName:          "dwh_img_tasks",
		ImgQueueMaxPriority:   10,
		ImgQueuePrefetchCount: 1,

		StoreAddr:           "http://localhost",
		StorePort:           DefaultStorePort,
		Resolutions:         []Resolution{{640, 480}, {440, 330}, {200, 150}, {120, 90}},
		InterpolationMethod: 2,
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

func ReadDwhImageServiceConfig(configName, path string) *DwhQueueServiceConfig {
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
