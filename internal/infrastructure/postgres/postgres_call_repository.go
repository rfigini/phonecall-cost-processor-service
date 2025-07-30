package postgres

import (
	"database/sql"
	"fmt"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/repository"
	"phonecall-cost-processor-service/internal/infrastructure/postgres/entity"
)

// PostgresCallRepository implementa repository.CallRepository usando DTOs de persistencia
type PostgresCallRepository struct {
	db *sql.DB
}

// NewPostgresCallRepository crea una nueva instancia con la conexión a la DB
func NewPostgresCallRepository(db *sql.DB) *PostgresCallRepository {
	return &PostgresCallRepository{db: db}
}

var _ repository.CallRepository = (*PostgresCallRepository)(nil)

// SaveIncomingCall persiste la llamada inicial con status=PENDING y processed_at
func (r *PostgresCallRepository) SaveIncomingCall(call model.NewIncomingCall) error {
	// Mapear dominio a entidad de persistencia
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

// UpdateCallCost actualiza el costo y marca status=COST_FETCHED con processed_at=NOW()
func (r *PostgresCallRepository) UpdateCallCost(callID string, cost float64, currency string) error {
	const query = `
	UPDATE calls
	SET cost = $1,
		currency = $2,
		status = 'COST_FETCHED',
		processed_at = NOW()
	WHERE call_id = $3
	AND status != 'REFUNDED';
	`
	if _, err := r.db.Exec(query, cost, currency, callID); err != nil {
		return fmt.Errorf("error actualizando costo: %w", err)
	}
	return nil
}

// MarkCostAsFailed marca status=COST_FETCH_FAILED con processed_at=NOW()
func (r *PostgresCallRepository) MarkCostAsFailed(callID string) error {
	const query = `
	UPDATE calls
	SET status = 'COST_FETCH_FAILED',
		processed_at = NOW()
	WHERE call_id = $1
	AND status != 'REFUNDED';
	`
	if _, err := r.db.Exec(query, callID); err != nil {
		return fmt.Errorf("error marcando fallo de costo: %w", err)
	}
	return nil
}

// ApplyRefund marca como refund y actualiza status=REFUNDED con processed_at=NOW()
func (r *PostgresCallRepository) ApplyRefund(refund model.RefundCall) error {
	// Mapear dominio a entidad
	e := entity.FromRefundCall(refund)

	const query = `
	INSERT INTO calls (
		call_id, refunded, refund_reason, cost, status, processed_at
	) VALUES ($1, true, $2, 0, 'REFUNDED', NOW())
	ON CONFLICT (call_id) DO UPDATE
	  SET refunded      = true,
	      refund_reason = EXCLUDED.refund_reason,
	      cost          = 0,
	      status        = 'REFUNDED',
	      processed_at  = NOW();
	`
	_, err := r.db.Exec(query, e.CallID, *e.RefundReason)
	if err != nil {
		return fmt.Errorf("error aplicando refund: %w", err)
	}
	return nil
}

// GetCallStatus devuelve el status de la llamada o vacío si no existe
func (r *PostgresCallRepository) GetCallStatus(callID string) (string, error) {
    const query = `SELECT status FROM calls WHERE call_id = $1`
    var status string
    err := r.db.QueryRow(query, callID).Scan(&status)
    if err == sql.ErrNoRows {
        return "", nil
    }
    return status, err
}


