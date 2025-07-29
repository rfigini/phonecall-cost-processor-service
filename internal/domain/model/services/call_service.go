package service

import (
	"log"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/client"
	"phonecall-cost-processor-service/internal/domain/port/repository"
)

type ICallService interface {
	Process(call model.NewIncomingCall) error
}

type CallService struct {
	repo       repository.CallRepository
	costClient client.CostClient
}

func NewCallService(repo repository.CallRepository, costClient client.CostClient) ICallService {
	return &CallService{repo, costClient}
}

func (s *CallService) Process(call model.NewIncomingCall) error {
	if err := s.repo.SaveIncomingCall(call); err != nil {
		return err
	}

	costResp, err := s.costClient.GetCallCost(call.CallID)
	if err != nil {
		log.Printf("⚠️ Error obteniendo costo: %v", err)
		call.CostFetchFailed = true
		// guardamos que falló la obtención del costo
		return s.repo.UpdateCallCost(call.CallID, 0, "")
	}

	return s.repo.UpdateCallCost(call.CallID, costResp.Cost, costResp.Currency)
}
