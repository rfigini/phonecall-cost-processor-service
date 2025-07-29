package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"phonecall-cost-processor-service/internal/model"
)

type CallRepository struct {
	db *sql.DB
}

func NewCallRepository(db *sql.DB) *CallRepository {
	return &CallRepository{db: db}
}

func (r *CallRepository) SaveIncomingCall(call model.NewIncomingCall) error {
	query := `
		INSERT INTO calls (
			call_id, caller, receiver, duration_in_seconds, start_timestamp, cost, currency, cost_fetch_failed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (call_id) DO NOTHING;
	`

	_, err := r.db.Exec(query,
		call.CallID,
		call.Caller,
		call.Receiver,
		call.DurationInSec,
		call.StartTimestamp,
		call.Cost,
		call.Currency,
		call.CostFetchFailed,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			fmt.Printf("⚠️ Llamada duplicada ignorada: %s\n", call.CallID)
			return nil
		}
		return fmt.Errorf("error insertando llamada: %w", err)
	}

	return nil
}

func (r *CallRepository) ApplyRefund(refund model.RefundCall) error {
	query := `
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

func (r *CallRepository) UpdateCallCost(callID string, cost float64, currency string) error {
	query := `
		UPDATE calls
		SET cost = $1,
		    currency = $2
		WHERE call_id = $3;
	`

	_, err := r.db.Exec(query, cost, currency, callID)
	if err != nil {
		return fmt.Errorf("error actualizando costo: %w", err)
	}

	return nil
}
