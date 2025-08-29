package bootstrap

import (
	"media-service/domain/usecase"
	"media-service/infrastructure/grpc_service"
	"media-service/infrastructure/repo"
	"time"

	"media-service/proto/media/v1"

	"github.com/anhvanhoa/service-core/bootstrap/db"
	grpc_server "github.com/anhvanhoa/service-core/bootstrap/grpc"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/queue"
	"github.com/go-pg/pg/v10"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App represents the application with all its dependencies
type App struct {
	Env            *Env
	DB             *pg.DB
	Logger         *log.LogGRPCImpl
	GRPCServer     *grpc.Server
	QueueClient    queue.QueueClient
	MediaUsecases  *usecase.MediaUsecases
	StorageService usecase.StorageService
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

	storageService := repo.NewLocalStorageService(
		env.STORAGE_LOCAL.UPLOAD_DIR,
		env.STORAGE_LOCAL.PUBLIC_URL,
		logger,
	)

	// processingService := repo.NewMediaProcessingService(
	// 	variantRepo,
	// 	storageService,
	// 	queueClient,
	// 	logger,
	// )

	mediaUsecases := usecase.NewMediaUsecases(
		mediaRepo,
		variantRepo,
		logger,
	)

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	mediaServiceServer := grpc_service.NewMediaServiceServer(mediaUsecases, logger)
	media.RegisterMediaServiceServer(grpcServer, mediaServiceServer)

	// Enable reflection for development
	if env.NODE_ENV == "development" {
		reflection.Register(grpcServer)
	}

	return &App{
		Env:            env,
		DB:             db,
		Logger:         logger,
		GRPCServer:     grpcServer,
		QueueClient:    queueClient,
		MediaUsecases:  mediaUsecases,
		StorageService: storageService,
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
			// media.RegisterMediaServiceServer(server, app.MediaUsecases)
		},
	)
}
