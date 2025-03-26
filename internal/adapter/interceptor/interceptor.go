package interceptor

import(
	"context"
	
	"github.com/rs/zerolog/log"
	"github.com/go-tokenization-grpc/internal/core/erro"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.adapter.interceptor").Logger()
var tracerProvider go_core_observ.TracerProvider

// About authentication intercetor
func ServerUnaryInterceptor(	ctx context.Context, 
								req any, _ *grpc.UnaryServerInfo, 
								handler grpc.UnaryHandler ) (any, error) {
	childLogger.Info().Str("func","ServerUnaryInterceptor").Send()

	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, erro.MissingData
	}

	childLogger.Info().Str("func","ServerUnaryInterceptor").Interface("headers",headers).Send()

	if len(headers["authorization"]) == 0 {
		childLogger.Info().Msg("WITHOUT AUTHORIZATION")
	}

	// Add trace into context
	if len(headers["trace-resquest-id"]) > 0 {	
		ctx = context.WithValue(ctx, "trace-request-id", headers["trace-resquest-id"][0] )
	}
	return handler(ctx, req)
}