package database

import (
	"time"
	"context"
	"errors"
	
	"github.com/go-tokenization-grpc/internal/core/erro"
	"github.com/go-tokenization-grpc/internal/core/model"
	"github.com/jackc/pgx/v5"
)

// About add payment
func (w *WorkerRepository) AddPayment(ctx context.Context, tx pgx.Tx, payment *model.Payment) (*model.Payment, error){
	childLogger.Info().Str("func","AddPayment").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	span := tracerProvider.Span(ctx, "database.AddPayment")
	defer span.End()

	// Prepare
	payment.CreatedAt = time.Now()
	if payment.PaymentAt.IsZero(){
		payment.PaymentAt = payment.CreatedAt
	}

	// Query and execute
	query := `INSERT INTO payment (fk_card_id, 
									card_number, 
									fk_terminal_id, 
									terminal, 
									card_type, 
									card_model, 
									payment_at, 
									mcc, 
									status, 
									currency, 
									amount, 
									created_at,
									tenant_id)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`

	row := tx.QueryRow(ctx, query, payment.FkCardId,
									payment.CardNumber,
									payment.FkTerminalId,
									payment.Terminal,
									payment.CardType,
									payment.CardModel,
									payment.PaymentAt,
									payment.Mcc,
									payment.Status,
									payment.Currency,
									payment.Amount,
									payment.CreatedAt ,
									payment.TenantID)

	var id int
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	// set PK
	payment.ID = id

	return payment , nil
}

// About get terminal
func (w *WorkerRepository) GetTerminal(ctx context.Context, terminal model.Terminal) (*model.Terminal, error){
	childLogger.Info().Str("func","GetTerminal").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.GetTerminal")
	defer span.End()

	// Get connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare
	res_terminal := model.Terminal{}

	// query and execute
	query :=  `SELECT 	id, 
						name, 
						coord_x, 
						coord_y, 
						status, 
						created_at, 
						updated_at
				FROM terminal
				WHERE name =$1`

	rows, err := conn.Query(ctx, query, terminal.Name)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_terminal.ID, 
							&res_terminal.Name, 
							&res_terminal.CoordX, 
							&res_terminal.CoordY, 
							&res_terminal.Status,
							&res_terminal.CreatedAt,
							&res_terminal.UpdatedAt,
		)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_terminal, nil
	}
	
	return nil, erro.ErrNotFound
}

// About update payment
func (w *WorkerRepository) UpdatePayment(ctx context.Context, tx pgx.Tx, payment model.Payment) (int64, error){
	childLogger.Info().Str("func","UpdatePayment").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	span := tracerProvider.Span(ctx, "database.UpdatePayment")
	defer span.End()

	// Query and execute
	query := `update payment
				set status = $2,
					updated_at = $3
				where id = $1`

	row, err := tx.Exec(ctx, query,	payment.ID,
									payment.Status,
									time.Now())
	if err != nil {
		return 0, errors.New(err.Error())
	}
	return row.RowsAffected(), nil
}