package app

import (
	"context"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	repository "urlshort/internal/repository/urlshort"

	urlshortGenerated "pkg/generated/urlshort/api/urlshort/v1"
	"urlshort/internal/config"
	"urlshort/internal/controller"
	"urlshort/internal/controller/urlshort"
	urlshortService "urlshort/internal/usecase/urlshort"
	"urlshort/migrations"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	const (
		GracefulShutdownTimeout = time.Second * 3
		POSTGRES                = "POSTGRES"
	)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var dbPool *pgxpool.Pool
	if cfg.PG.AccessType == POSTGRES {
		var err error

		dbPool, err = pgxpool.New(ctx, cfg.ConstructPostgresURL())
		if err != nil {
			logger.Error("failed to connect to database", zap.Error(err))
			return
		}
		defer dbPool.Close()

		err = migrations.SetupPostgres(dbPool, logger)
		if err != nil {
			logger.Error("failed to setup migrations", zap.Error(err))
			return
		}
	}

	repo, err := repository.New(logger, dbPool, cfg.PG.AccessType)
	if err != nil {
		logger.Error("failed to setup repository", zap.Error(err))
		return
	}

	urlshortUsecase := urlshortService.NewURLShortService(repo)

	ctrlURLShort := urlshort.NewURLShortServer(logger, urlshortUsecase)

	ctrl := controller.New(ctrlURLShort)

	go runGrpc(logger, cfg, ctrl)
	go runRest(ctx, logger, cfg)

	<-ctx.Done()
	time.Sleep(GracefulShutdownTimeout)
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := urlshortGenerated.RegisterUrlshortHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway", zap.Error(err))
		return
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	logger.Info("gateway listening at port", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
		return
	}
	server := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			),
		),
	)
	reflection.Register(server)
	urlshortGenerated.RegisterUrlshortServer(server, ctrl)

	logger.Info("grpc server listening at port", zap.String("port", cfg.GRPC.Port))

	if err = server.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}
