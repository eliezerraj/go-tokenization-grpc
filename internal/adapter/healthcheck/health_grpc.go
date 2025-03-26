package healthcheck

import (
	"context"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
	"github.com/rs/zerolog/log"
  )


var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.adapter.healthcheck").Logger()
var startTime = time.Now()

type HealthChecker struct{}

func NewHealthChecker() *HealthChecker {
	childLogger.Info().Str("func","NewHealthChecker").Send()
	
	return &HealthChecker{}
}

// About check
func (s *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	childLogger.Info().Str("func","Check").Send()

	//var currentTime = time.Now()
	var currentStatus = grpc_health_v1.HealthCheckResponse_SERVING
	// simulating unavailability ater two minutes
	/*if currentTime.Sub(startTime).Minutes() > 2 {
		currentStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}*/
	health_check_response := &grpc_health_v1.HealthCheckResponse{
		Status: currentStatus,
	}
	return health_check_response, nil
}

// About Watch
func (s *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	childLogger.Info().Str("func","Watch").Send()

	//var currentTime = time.Now()
	var currentStatus = grpc_health_v1.HealthCheckResponse_SERVING
	// simulating unavailability ater two minutes
	/*if currentTime.Sub(startTime).Minutes() > 2 {
		currentStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}*/
	health_check_response := &grpc_health_v1.HealthCheckResponse{
		Status: currentStatus,
	}
	return server.Send(health_check_response)
}

