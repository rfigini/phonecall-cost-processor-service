package client

import (
	"phonecall-cost-processor-service/internal/domain/model"
)

type CostClient interface {
	GetCallCost(callID string) (*model.CostResponse, error)
}
