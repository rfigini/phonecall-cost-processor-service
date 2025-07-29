package model

type CostResponse struct {
	Currency string  `json:"currency"`
	Cost     float64 `json:"cost"`
}
