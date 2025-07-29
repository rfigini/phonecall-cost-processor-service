package entity

import (
	"fmt"
	"time"

	"phonecall-cost-processor-service/internal/domain/model"
)

// CallEntity representa la fila de la tabla 'calls' en Postgres
// con tags para mapeo SQL.
type CallEntity struct {
	CallID          string     `db:"call_id"`
	Caller          string     `db:"caller"`
	Receiver        string     `db:"receiver"`
	DurationInSec   int        `db:"duration_in_seconds"`
	StartTimestamp  time.Time  `db:"start_timestamp"`
	Cost            float64    `db:"cost"`
	Currency        string     `db:"currency"`
	CostFetchFailed bool       `db:"cost_fetch_failed"`
	Refunded        bool       `db:"refunded"`
	RefundReason    *string    `db:"refund_reason"`
	ProcessedAt     time.Time  `db:"processed_at"`
}

// FromNewIncomingCall mapea el modelo de dominio NewIncomingCall a CallEntity.
// Parsea el timestamp RFC3339 y retorna error si es inválido.
func FromNewIncomingCall(m model.NewIncomingCall) (CallEntity, error) {
	// Parsear timestamp
	ts, err := time.Parse(time.RFC3339, m.StartTimestamp)
	if err != nil {
		return CallEntity{}, fmt.Errorf("start_timestamp inválido: %w", err)
	}

	return CallEntity{
		CallID:         m.CallID,
		Caller:         m.Caller,
		Receiver:       m.Receiver,
		DurationInSec:  m.DurationInSec,
		StartTimestamp: ts,
	}, nil
}

// FromRefundCall mapea el modelo de dominio RefundCall a CallEntity
// Prepara los campos de refund (refunded, refund_reason y reset de cost).
func FromRefundCall(m model.RefundCall) CallEntity {
	reason := m.Reason
	return CallEntity{
		CallID:       m.CallID,
		Refunded:     true,
		RefundReason: &reason,
		Cost:         0,
	}
}
