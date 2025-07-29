package dto

type NewIncomingCallDTO struct {
  CallID         string `json:"call_id"`
  Caller         string `json:"caller"`
  Receiver       string `json:"receiver"`
  DurationInSec  int    `json:"duration_in_seconds"`
  StartTimestamp string `json:"start_timestamp"`
}

type RefundCallDTO struct {
  CallID string `json:"call_id"`
  Reason string `json:"reason"`
}

