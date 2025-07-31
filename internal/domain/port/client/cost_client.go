package client

import (
	"fmt"
	"phonecall-cost-processor-service/internal/domain/model"
)

type CostClient interface {
	GetCallCost(callID string) (*model.CostResponse, error)
}

type CostAPIError struct {
	StatusCode int
	Err        error
}

func (e *CostAPIError) Error() string {
	return fmt.Sprintf("status code %d: %v", e.StatusCode, e.Err)
}

func (e *CostAPIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}