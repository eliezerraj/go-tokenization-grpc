package service

import(
	"context"

	"github.com/go-tokenization-grpc/internal/adapter/database"
	"github.com/go-tokenization-grpc/internal/core/model"
	"github.com/rs/zerolog/log"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.core.service").Logger()
var tracerProvider go_core_observ.TracerProvider

type WorkerService struct {
	apiService		[]model.ApiService
	workerRepository *database.WorkerRepository
}

// About create a new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository,
						apiService		[]model.ApiService) *WorkerService{

	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		apiService: apiService,
	}
}

// About get the card from token
func (s * WorkerService) GetCardToken(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardToken").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetCardToken")
	defer span.End()

	// Get card token information from repo
	res, err := s.workerRepository.GetCardToken(ctx, card)
	if err != nil {
		return nil, err
	}

	return res, nil
}
