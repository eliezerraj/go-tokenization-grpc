package database

import (
	"context"
	"errors"
	
	"github.com/go-tokenization-grpc/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.adapter.database").Logger()

var tracerProvider go_core_observ.TracerProvider

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

// About create a worker repository
func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

// About add token card 
func (w *WorkerRepository) GetCardToken(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardToken").Interface("trace-request-id", ctx.Value("trace-request-id")).Send()

	//trace
	span := tracerProvider.Span(ctx, "database.GetCardToken")
	defer span.End()

	// Prepare
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_card := model.Card{}
	res_card_list := []model.Card{}
	
	// Query e Execute
	query := `SELECT ca.id,
					ca.fk_account_id,
					ca.card_number,
					ca.card_type,
					ca.card_model, 
					ct.token,
					ca.atc,
					ct.status,
					ct.expired_at,
					ct.created_at,
					ct.updated_at,																									
					ct.tenant_id	
			FROM card_token ct,
				card ca
			WHERE ct.token = $1
			and ca.id = ct.fk_id_card
			order by ct.created_at desc`

	rows, err := conn.Query(ctx, query, string(card.TokenData))
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_card.ID,
							&res_card.FkAccountId, 
							&res_card.CardNumber,
							&res_card.Type,
							&res_card.Model, 
							&res_card.TokenData,
							&res_card.Atc, 
							&res_card.Status,
							&res_card.ExpiredAt,
							&res_card.CreatedAt,
							&res_card.UpdatedAt,
							&res_card.TenantID)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_card_list = append(res_card_list, res_card)
	}

	return &res_card_list , nil
}
