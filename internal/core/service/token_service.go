package service

import(
	"fmt"
	"time"
	"context"
	"github.com/zeebo/blake3"

	"github.com/go-tokenization-grpc/internal/adapter/database"
	"github.com/go-tokenization-grpc/internal/core/model"
	"github.com/go-tokenization-grpc/internal/core/erro"
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

	card.CreatedAt = time.Now()
	card.ExpiredAt = time.Now().AddDate(0, 3, 0) // Add 3 months

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

// About create a tokenization data
func (s * WorkerService) AddPaymentToken(ctx context.Context, payment model.Payment) (*model.Payment, error){
	childLogger.Info().Str("func","AddPaymentToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("payment", payment).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.AddPaymentToken")

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

	// Businness rule
	if (payment.CardType != "CREDIT") && (payment.CardType != "DEBIT") {
		return nil, erro.ErrCardTypeInvalid
	}

	// Get terminal
	terminal := model.Terminal{Name: payment.Terminal}
	res_terminal, err := s.workerRepository.GetTerminal(ctx, terminal)
	if err != nil {
		return nil, err
	}

	// Get Card data
	card := model.Card{TokenData: payment.TokenData}
	res_list_card, err := s.workerRepository.GetCardToken(ctx, card)

	if err != nil {
		return nil, err
	}
	if len(*res_list_card) == 0 {
		return nil, erro.ErrNotFound
	}

	// Prepare payment
	payment.FkCardId = (*res_list_card)[0].ID
	payment.CardNumber = (*res_list_card)[0].CardNumber
	payment.CardModel = (*res_list_card)[0].Model
	payment.FkTerminalId = res_terminal.ID
	payment.Status = "AUTHORIZATION-PENDING:GRPC"

	res_payment, err := s.workerRepository.AddPayment(ctx, tx, &payment)
	if err != nil {
		return nil, err
	}

	// update status payment
	res_update, err := s.workerRepository.UpdatePayment(ctx, tx, *res_payment)
	if err != nil {
		return nil, err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return nil, err
	}

	return res_payment, nil
}
