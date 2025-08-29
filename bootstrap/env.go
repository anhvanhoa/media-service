package bootstrap

import (
	"strings"

	"github.com/anhvanhoa/service-core/bootstrap/config"
	"github.com/anhvanhoa/service-core/domain/grpc_client"
)

type StorageLocal struct {
	UPLOAD_DIR string `mapstructure:"upload_dir"`
	PUBLIC_URL string `mapstructure:"public_url"`
}

type QueueRedis struct {
	ADDR     string `mapstructure:"addr"`
	DB       int    `mapstructure:"db"`
	PASSWORD string `mapstructure:"password"`
	NETWORK  string `mapstructure:"network"`
	TIMEOUT  int    `mapstructure:"timeout"`
	TLS      bool   `mapstructure:"tls"`
	RETRY    int    `mapstructure:"retry"`
}

type Env struct {
	NODE_ENV       string `mapstructure:"node_env"`
	URL_DB         string `mapstructure:"url_db"`
	NAME_SERVICE   string `mapstructure:"name_service"`
	PORT_GRPC      int    `mapstructure:"port_grpc"`
	HOST_GRPC      string `mapstructure:"host_grpc"`
	INTERVAL_CHECK string `mapstructure:"interval_check"`
	TIMEOUT_CHECK  string `mapstructure:"timeout_check"`

	QUEUE *QueueRedis `mapstructure:"queue"`

	STORAGE_LOCAL *StorageLocal `mapstructure:"storage_local"`

	GRPC_CLIENTS []*grpc_client.ConfigGrpc `mapstructure:"grpc_clients"`
}

func NewEnv(env any) {
	setting := config.DefaultSettingsConfig()
	if setting.IsProduction() {
		setting.SetFile("prod.config")
	} else {
		setting.SetFile("dev.config")
	}
	config.NewConfig(setting, env)
}

func (env *Env) IsProduction() bool {
	return strings.ToLower(env.NODE_ENV) == "production"
}
