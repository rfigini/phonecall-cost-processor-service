package entity

import (
	"time"

	"phonecall-cost-processor-service/internal/domain/model"
)

type CallEntity struct {
	CallID         string    `db:"call_id"`
	Caller         string    `db:"caller"`
	Receiver       string    `db:"receiver"`
	DurationInSec  int       `db:"duration_in_seconds"`
	StartTimestamp time.Time `db:"start_timestamp"`
	Cost           float64   `db:"cost"`
	Currency       string    `db:"currency"`
	Refunded       bool      `db:"refunded"`
	RefundReason   *string   `db:"refund_reason"`
	Status         string    `db:"status"`
	ProcessedAt    time.Time `db:"processed_at"`
}

func FromNewIncomingCall(m model.NewIncomingCall) (CallEntity, error) {
ts, err := time.Parse(time.RFC3339, m.StartTimestamp)
	if err != nil {
		return CallEntity{}, err
	}
	return CallEntity{
		CallID:         m.CallID,
		Caller:         m.Caller,
		Receiver:       m.Receiver,
		DurationInSec:  m.DurationInSec,
		StartTimestamp: ts,
		Status:         "PENDING",
		ProcessedAt:    time.Now(),
	}, nil
}

func FromRefundCall(m model.RefundCall) CallEntity {
	reason := m.Reason
	return CallEntity{
		CallID:       m.CallID,
		Refunded:     true,
		RefundReason: &reason,
		Status:       "REFUNDED",
		ProcessedAt:  time.Now(),
	}
}
