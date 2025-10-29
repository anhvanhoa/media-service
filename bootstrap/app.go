package bootstrap

import (
	"media-service/domain/usecase"
	"media-service/infrastructure/grpc_service"
	"media-service/infrastructure/repo"

	"github.com/anhvanhoa/sf-proto/gen/media/v1"

	"github.com/anhvanhoa/service-core/bootstrap/db"
	grpc_server "github.com/anhvanhoa/service-core/bootstrap/grpc"
	"github.com/anhvanhoa/service-core/domain/cache"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/queue"
	"github.com/anhvanhoa/service-core/domain/storage"
	"github.com/anhvanhoa/service-core/domain/token"
	"github.com/anhvanhoa/service-core/domain/user_context"
	"github.com/anhvanhoa/service-core/utils"
	"github.com/go-pg/pg/v10"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type App struct {
	Env           *Env
	DB            *pg.DB
	Logger        *log.LogGRPCImpl
	GRPCServer    *grpc.Server
	QueueClient   queue.QueueClient
	MediaUsecases usecase.MediaUsecaseInterfaces
	Storage       storage.StorageI
	MediaServer   media.MediaServiceServer
	Helper        utils.Helper
	Cache         cache.CacheI
}

func NewApp() *App {
	env := &Env{}
	NewEnv(env)

	logConfig := log.NewConfig()
	logger := log.InitLogGRPC(logConfig, zapcore.DebugLevel, env.IsProduction())

	db := db.NewPostgresDB(db.ConfigDB{
		URL:  env.UrlDb,
		Mode: env.NodeEnv,
	})

	queueClient := queue.NewQueueClient(queue.NewDefaultConfig(
		env.Queue.Addr,
		env.Queue.Network,
		env.Queue.Password,
		env.Queue.Db,
		nil,
		env.Queue.Retry,
	))
	mediaRepo := repo.NewMediaRepository(db)

	storageService := storage.NewLocalStorageService(
		env.StorageLocal.UploadDir,
		logger,
	)

	processingService := processing.NewMediaProcessingService(
		storageService,
		queueClient,
		logger,
	)

	mediaUsecases := usecase.NewMediaUsecases(
		mediaRepo,
		logger,
		processingService,
		storageService,
	)

	helper := utils.NewHelper()
	configRedis := cache.NewConfigCache(
		env.DbCache.Addr,
		env.DbCache.Password,
		env.DbCache.Db,
		env.DbCache.Network,
		env.DbCache.MaxIdle,
		env.DbCache.MaxActive,
		env.DbCache.IdleTimeout,
	)
	cache := cache.NewCache(configRedis)

	mediaServiceServer := grpc_service.NewMediaServiceServer(mediaUsecases, logger)

	return &App{
		Env:           env,
		DB:            db,
		Logger:        logger,
		QueueClient:   queueClient,
		MediaUsecases: mediaUsecases,
		Storage:       storageService,
		MediaServer:   mediaServiceServer,
		Helper:        helper,
		Cache:         cache,
	}
}

func (app *App) Start() *grpc_server.GRPCServer {
	config := &grpc_server.GRPCServerConfig{
		IsProduction: app.Env.IsProduction(),
		PortGRPC:     app.Env.PortGrpc,
		NameService:  app.Env.NameService,
	}
	middleware := grpc_server.NewMiddleware(
		token.NewToken(app.Env.AccessSecret),
	)
	return grpc_server.NewGRPCServer(
		config,
		app.Logger,
		func(server *grpc.Server) {
			media.RegisterMediaServiceServer(server, app.MediaServer)
		},
		middleware.AuthorizationInterceptor(
			app.Env.SecretService,
			func(action string, resource string) bool {
				hasPermission, err := app.Cache.Get(resource + "." + action)
				if err != nil {
					return false
				}
				return hasPermission != nil && string(hasPermission) == "true"
			},
			func(id string) *user_context.UserContext {
				userData, err := app.Cache.Get(id)
				if err != nil || userData == nil {
					return nil
				}
				uCtx := user_context.NewUserContext()
				uCtx.FromBytes(userData)
				return uCtx
			},
		),
	)
}
