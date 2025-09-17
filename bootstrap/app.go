package bootstrap

import (
	"media-service/domain/usecase"
	"media-service/infrastructure/grpc_service"
	"media-service/infrastructure/repo"

	"github.com/anhvanhoa/sf-proto/gen/media/v1"

	"github.com/anhvanhoa/service-core/bootstrap/db"
	grpc_server "github.com/anhvanhoa/service-core/bootstrap/grpc"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/queue"
	"github.com/anhvanhoa/service-core/domain/storage"
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

	mediaServiceServer := grpc_service.NewMediaServiceServer(mediaUsecases, logger)

	return &App{
		Env:           env,
		DB:            db,
		Logger:        logger,
		QueueClient:   queueClient,
		MediaUsecases: mediaUsecases,
		Storage:       storageService,
		MediaServer:   mediaServiceServer,
	}
}

func (app *App) Start() *grpc_server.GRPCServer {
	config := &grpc_server.GRPCServerConfig{
		IsProduction: app.Env.IsProduction(),
		PortGRPC:     app.Env.PortGrpc,
		NameService:  app.Env.NameService,
	}
	return grpc_server.NewGRPCServer(
		config,
		app.Logger,
		func(server *grpc.Server) {
			media.RegisterMediaServiceServer(server, app.MediaServer)
		},
	)
}
