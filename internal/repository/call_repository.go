package repository

import (
	"database/sql"
	"fmt"

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
			call_id, caller, receiver, duration_in_seconds, start_timestamp
		)
		VALUES ($1, $2, $3, $4, $5)
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

