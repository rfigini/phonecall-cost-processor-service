package services

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
	return &CallService{repo: repo, costClient: costClient}
}

func (s *CallService) Process(call model.NewIncomingCall) error {
	status, err := s.repo.GetCallStatus(call.CallID)
	if err != nil {
		return err
	}
	if status == "REFUNDED" {
		log.Printf("⚠️ Esta llamada ya estaba refundeada call_id=%s: %v", call.CallID, err)
		return nil
	}

	if err := s.repo.SaveIncomingCall(call); err != nil {
		return err
	}

	costResp, err := s.costClient.GetCallCost(call.CallID)
	if err != nil {
		log.Printf("⚠️ Error obteniendo costo para call_id=%s: %v", call.CallID, err)
		return s.repo.MarkCostAsFailed(call.CallID)
	}

	return s.repo.UpdateCallCost(call.CallID, costResp.Cost, costResp.Currency)
}
