package bootstrap

import (
	"media-service/domain/usecase"
	"media-service/infrastructure/grpc_service"
	"media-service/infrastructure/repo"
	"media-service/proto/media/v1"
	"time"

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

// App represents the application with all its dependencies
type App struct {
	Env           *Env
	DB            *pg.DB
	Logger        *log.LogGRPCImpl
	GRPCServer    *grpc.Server
	QueueClient   queue.QueueClient
	MediaUsecases *usecase.MediaUsecases
	Storage       storage.StorageI
	MediaServer   media.MediaServiceServer
}

func NewApp() *App {
	env := &Env{}
	NewEnv(env)

	logConfig := log.NewConfig()
	logger := log.InitLogGRPC(logConfig, zapcore.DebugLevel, env.IsProduction())

	// Initialize database
	db := db.NewPostgresDB(db.ConfigDB{
		URL:  env.URL_DB,
		Mode: env.NODE_ENV,
	})

	queueClient := queue.NewQueueClient(queue.NewDefaultConfig(
		env.QUEUE.ADDR,
		env.QUEUE.NETWORK,
		env.QUEUE.PASSWORD,
		env.QUEUE.DB,
		time.Duration(env.QUEUE.TIMEOUT),
		nil,
		env.QUEUE.RETRY,
	))
	mediaRepo := repo.NewMediaRepository(db)
	variantRepo := repo.NewMediaVariantRepository(db)

	storageService := storage.NewLocalStorageService(
		env.STORAGE_LOCAL.UPLOAD_DIR,
		logger,
	)

	processingService := processing.NewMediaProcessingService(
		storageService,
		queueClient,
		logger,
	)

	mediaUsecases := usecase.NewMediaUsecases(
		mediaRepo,
		variantRepo,
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
		PortGRPC:     app.Env.PORT_GRPC,
		NameService:  app.Env.NAME_SERVICE,
	}
	return grpc_server.NewGRPCServer(
		config,
		app.Logger,
		func(server *grpc.Server) {
			media.RegisterMediaServiceServer(server, app.MediaServer)
		},
	)
}
