package bootstrap

import (
	"strings"

	"github.com/anhvanhoa/service-core/bootstrap/config"
	"github.com/anhvanhoa/service-core/domain/grpc_client"
)

type StorageLocal struct {
	UploadDir string `mapstructure:"upload_dir"`
}

type QueueRedis struct {
	Addr     string `mapstructure:"addr"`
	Db       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
	Network  string `mapstructure:"network"`
	Timeout  int    `mapstructure:"timeout"`
	Tls      bool   `mapstructure:"tls"`
	Retry    int    `mapstructure:"retry"`
}
type dbCache struct {
	Addr        string `mapstructure:"addr"`
	Db          int    `mapstructure:"db"`
	Password    string `mapstructure:"password"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxActive   int    `mapstructure:"max_active"`
	IdleTimeout int    `mapstructure:"idle_timeout"`
	Network     string `mapstructure:"network"`
}

type Env struct {
	NodeEnv               string                    `mapstructure:"node_env"`
	SecretService         string                    `mapstructure:"secret_service"`
	UrlDb                 string                    `mapstructure:"url_db"`
	NameService           string                    `mapstructure:"name_service"`
	PortGrpc              int                       `mapstructure:"port_grpc"`
	HostGrpc              string                    `mapstructure:"host_grpc"`
	IntervalCheck         string                    `mapstructure:"interval_check"`
	TimeoutCheck          string                    `mapstructure:"timeout_check"`
	Queue                 *QueueRedis               `mapstructure:"queue"`
	StorageLocal          *StorageLocal             `mapstructure:"storage_local"`
	GrpcClients           []*grpc_client.ConfigGrpc `mapstructure:"grpc_clients"`
	PermissionServiceAddr string                    `mapstructure:"permission_service_addr"`
	DbCache               *dbCache                  `mapstructure:"db_cache"`
}

func NewEnv(env any) {
	setting := config.DefaultSettingsConfig()
	if setting.IsProduction() {
		setting.SetPath("/config")
		setting.SetFile("media_service.config")
	} else {
		setting.SetFile("dev.config")
	}
	config.NewConfig(setting, env)
}

func (env *Env) IsProduction() bool {
	return strings.ToLower(env.NodeEnv) == "production"
}
