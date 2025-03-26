package service

import(
	"fmt"
	"time"
	"context"
	"github.com/zeebo/blake3"

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

// About create a tokenization data
func (s * WorkerService) CreateCardToken(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.CreateCardToken")

	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// prepare data
	hasher := blake3.New()
	hasher.Write([]byte(card.CardNumber))
	card.TokenData = fmt.Sprintf("%x", (hasher.Sum(nil)) )
	card.Status = "ACTIVE"

	card.CreateAt = time.Now()
	card.ExpireAt = time.Now().AddDate(0, 3, 0) // Add 3 months

	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Call a service
	res, err := s.workerRepository.CreateCardToken(ctx, tx, card)
	if err != nil {
		return nil, err
	}

	// Setting PK
	card.ID = res.ID

	return &card, nil
}

// About get the card from token
func (s * WorkerService) GetCardToken(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.GetCardToken")
	defer span.End()

	// Call a service
	res, err := s.workerRepository.GetCardToken(ctx, card)
	if err != nil {
		return nil, err
	}

	return res, nil
}