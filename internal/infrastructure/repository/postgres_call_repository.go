package repository

import (
	"database/sql"
	"fmt"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/repository"
)

type PostgresCallRepository struct {
	db *sql.DB
}

func NewPostgresCallRepository(db *sql.DB) *PostgresCallRepository {
	return &PostgresCallRepository{db: db}
}

var _ repository.CallRepository = (*PostgresCallRepository)(nil)

func (r *PostgresCallRepository) SaveIncomingCall(call model.NewIncomingCall) error {
	const query = `
		INSERT INTO calls (
			call_id, caller, receiver, duration_in_seconds, start_timestamp
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (call_id) DO NOTHING;
	`

	_, err := r.db.Exec(query,
		call.CallID,
		call.Caller,
		call.Receiver,
		call.DurationInSec,
		call.StartTimestamp,
	)

	if err != nil {
		return fmt.Errorf("error insertando llamada: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) UpdateCallCost(callID string, cost float64, currency string) error {
	const query = `
		UPDATE calls
		SET cost = $1,
			currency = $2,
			cost_fetch_failed = false
		WHERE call_id = $3;
	`

	_, err := r.db.Exec(query, cost, currency, callID)
	if err != nil {
		return fmt.Errorf("error actualizando costo: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) MarkCostAsFailed(callID string) error {
	const query = `
		UPDATE calls
		SET cost_fetch_failed = true
		WHERE call_id = $1;
	`

	_, err := r.db.Exec(query, callID)
	if err != nil {
		return fmt.Errorf("error marcando fallo de costo: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) ApplyRefund(refund model.RefundCall) error {
	const query = `
		UPDATE calls
		SET refunded = true,
			refund_reason = $1,
			cost = 0
		WHERE call_id = $2;
	`

	_, err := r.db.Exec(query, refund.Reason, refund.CallID)
	if err != nil {
		return fmt.Errorf("error aplicando refund: %w", err)
	}
	return nil
}
