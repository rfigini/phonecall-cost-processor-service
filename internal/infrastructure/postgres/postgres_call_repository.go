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

// SaveIncomingCall mapea dominio -> entidad y persiste la llamada, sin duplicados
func (r *PostgresCallRepository) SaveIncomingCall(call model.NewIncomingCall) error {
	// 1) Mapear dominio a entidad
	e, err := entity.FromNewIncomingCall(call)
	if err != nil {
		return fmt.Errorf("error mapeando nueva llamada: %w", err)
	}

	// 2) Ejecutar INSERT ON CONFLICT
	const query = `
	INSERT INTO calls (
		call_id, caller, receiver, duration_in_seconds, start_timestamp
	) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (call_id) DO NOTHING;
	`
	if _, err := r.db.Exec(query,
		e.CallID,
		e.Caller,
		e.Receiver,
		e.DurationInSec,
		e.StartTimestamp,
	); err != nil {
		return fmt.Errorf("error insertando llamada: %w", err)
	}
	return nil
}

// UpdateCallCost actualiza el costo y marca el fetch como exitoso
func (r *PostgresCallRepository) UpdateCallCost(callID string, cost float64, currency string) error {
	const query = `
	UPDATE calls
	SET cost = $1,
		currency = $2,
		cost_fetch_failed = false
	WHERE call_id = $3;
	`
	if _, err := r.db.Exec(query, cost, currency, callID); err != nil {
		return fmt.Errorf("error actualizando costo: %w", err)
	}
	return nil
}

// MarkCostAsFailed marca que la obtención del costo falló
func (r *PostgresCallRepository) MarkCostAsFailed(callID string) error {
	const query = `
	UPDATE calls
	SET cost_fetch_failed = true
	WHERE call_id = $1;
	`
	if _, err := r.db.Exec(query, callID); err != nil {
		return fmt.Errorf("error marcando fallo de costo: %w", err)
	}
	return nil
}

// ApplyRefund mapea el refund a entidad y persiste el cambio
func (r *PostgresCallRepository) ApplyRefund(refund model.RefundCall) error {
	// 1) Mapear dominio a entidad
	e := entity.FromRefundCall(refund)

	const query = `
	UPDATE calls
	SET refunded = true,
		refund_reason = $1,
		cost = 0
	WHERE call_id = $2;
	`
	if _, err := r.db.Exec(query, e.RefundReason, e.CallID); err != nil {
		return fmt.Errorf("error aplicando refund: %w", err)
	}
	return nil
}
