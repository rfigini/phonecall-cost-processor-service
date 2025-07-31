package postgres

import (
	"database/sql"
	"fmt"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/repository"
	"phonecall-cost-processor-service/internal/infrastructure/postgres/entity"
)

type PostgresCallRepository struct {
	db *sql.DB
}

func (r *PostgresCallRepository) MarkCallAsInvalid(callID string) error {
	const query = `
		UPDATE calls
		SET status = 'INVALID',
			processed_at = NOW()
		WHERE call_id = $1;
	`
	_, err := r.db.Exec(query, callID)
	return err
}


func NewPostgresCallRepository(db *sql.DB) *PostgresCallRepository {
	return &PostgresCallRepository{db: db}
}

var _ repository.CallRepository = (*PostgresCallRepository)(nil)

func (r *PostgresCallRepository) SaveIncomingCall(call model.NewIncomingCall) error {
	e, err := entity.FromNewIncomingCall(call)
	if err != nil {
		return fmt.Errorf("error mapeando NewIncomingCall: %w", err)
	}

	const query = `
	INSERT INTO calls (
		call_id, caller, receiver, duration_in_seconds, start_timestamp, status, processed_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (call_id) DO NOTHING;
	`
	if _, err := r.db.Exec(query,
		e.CallID,
		e.Caller,
		e.Receiver,
		e.DurationInSec,
		e.StartTimestamp,
		e.Status,
		e.ProcessedAt,
	); err != nil {
		return fmt.Errorf("error insertando llamada: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) UpdateCallCost(callID string, cost float64, currency string) error {
	const query = `
	UPDATE calls
	SET cost = $1,
		currency = $2,
		status = 'OK',
		processed_at = NOW()
	WHERE call_id = $3
	AND status != 'REFUNDED';
	`
	if _, err := r.db.Exec(query, cost, currency, callID); err != nil {
		return fmt.Errorf("error actualizando costo: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) MarkCostAsFailed(callID string) error {
	const query = `
	UPDATE calls
	SET status = 'ERROR',
		processed_at = NOW()
	WHERE call_id = $1
	AND status != 'REFUNDED';
	`
	if _, err := r.db.Exec(query, callID); err != nil {
		return fmt.Errorf("error marcando fallo de costo: %w", err)
	}
	return nil
}

func (r *PostgresCallRepository) ApplyRefund(refund model.RefundCall) error {
	e := entity.FromRefundCall(refund)

	const query = `
	INSERT INTO calls (call_id, refunded, refund_reason, cost, status, processed_at)
	VALUES ($1, true, $2, 0, 'REFUNDED', NOW())
	ON CONFLICT (call_id) DO UPDATE
	SET refunded = true,
		refund_reason = EXCLUDED.refund_reason,
		cost = 0,
		status = 'REFUNDED',
		processed_at = NOW();
	`

	if _, err := r.db.Exec(query, e.CallID, e.RefundReason); err != nil {
		return fmt.Errorf("error aplicando refund: %w", err)
	}

	return nil
}

func (r *PostgresCallRepository) GetCallStatus(callID string) (string, error) {
	const query = `SELECT status FROM calls WHERE call_id = $1`
	var status string
	err := r.db.QueryRow(query, callID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return status, err
}

func (r *PostgresCallRepository) FillMissingCallData(call model.NewIncomingCall) error {
	const query = `
	UPDATE calls
	SET caller = $1,
		receiver = $2,
		duration_in_seconds = $3,
		start_timestamp = $4
	WHERE call_id = $5 AND status = 'REFUNDED';`
	_, err := r.db.Exec(query, call.Caller, call.Receiver, call.DurationInSec, call.StartTimestamp, call.CallID)
	return err
}
