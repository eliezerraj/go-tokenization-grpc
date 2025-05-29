package server

import (
	"time"
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc/health/grpc_health_v1"	
	"github.com/go-tokenization-grpc/internal/adapter/interceptor"
	"github.com/go-tokenization-grpc/internal/adapter/healthcheck"

	"github.com/go-tokenization-grpc/internal/core/model"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	token_proto_service "github.com/go-tokenization-grpc/protogen/token"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	grpc_adapter "github.com/go-tokenization-grpc/internal/adapter/grpc/server"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.infra.server").Logger()

var adapterGrpc grpc_adapter.AdapterGrpc
var tracer trace.Tracer

type WorkerServer struct {
	adapterGrpc *grpc_adapter.AdapterGrpc
}

// About create worker server
func NewWorkerServer(adapterGrpc *grpc_adapter.AdapterGrpc) *WorkerServer {
	childLogger.Info().Str("func","NewWorkerServer").Send()

	return &WorkerServer{
		adapterGrpc: adapterGrpc,
	}
}

// About start server
func (w *WorkerServer) StartGrpcServer(	ctx context.Context, 
										appServer *model.AppServer){
	childLogger.Info().Str("func","StartGrpcServer").Send()

	//Otel
	traceExporter, err := otlptracegrpc.New(ctx, 
											otlptracegrpc.WithInsecure(),
											otlptracegrpc.WithEndpoint(appServer.ConfigOTEL.OtelExportEndpoint),
											)
	if err != nil {
		childLogger.Error().Err(err).Msg("erro otlptracegrpc")
	}
	idg := xray.NewIDGenerator()

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(appServer.InfoPod.PodName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})
	tracer = otel.Tracer(appServer.InfoPod.PodName)

	// create grpc listener
	listener, err := net.Listen("tcp", appServer.Server.Port)
	if err != nil {
		childLogger.Error().Err(err).Msg("fatal error open service grpc")
		panic(err)
	}

	// prepare the options
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor( otelgrpc.UnaryServerInterceptor( ), interceptor.ServerUnaryInterceptor))
	opts = append(opts, grpc.KeepaliveParams(	keepalive.ServerParameters {
												MaxConnectionAge: time.Second * 30,
												MaxConnectionAgeGrace: time.Second * 10,
											}))
	
	// setup and prepare grpc server
	workerGrpcServer := grpc.NewServer(opts...)

	// handle defer
	defer func() {
		err = tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Send()
		}

		childLogger.Info().Msg("stopping server...")
		workerGrpcServer.Stop()
	
		childLogger.Info().Msg("stopping listener...")
		listener.Close()

		childLogger.Info().Msg("server stoped !!!")
	}()

	// wire
	token_proto_service.RegisterTokenServiceServer(workerGrpcServer, w.adapterGrpc)

	// health check
	healthService := healthcheck.NewHealthChecker()
	grpc_health_v1.RegisterHealthServer(workerGrpcServer, healthService)

	// run server
	go func(){
		childLogger.Info().Str("Service Port:", appServer.Server.Port).Send()
		
		if err := workerGrpcServer.Serve(listener); err != nil {
			childLogger.Error().Err(err).Msg("Failed to server!!!")
		}
	}()

	// Get SIGNALS
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGHUP:
			childLogger.Info().Msg("Received SIGHUP: reloading configuration...")
		case syscall.SIGINT, syscall.SIGTERM:
			childLogger.Info().Msg("Received SIGINT/SIGTERM termination signal. Exiting")
			return
		default:
			childLogger.Info().Interface("Received signal:", sig).Send()
		}
	}
}