package services

import (
	"errors"
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
	if status == "REFUND_PARTIALLY" {
		log.Printf("🔄 Completando datos de llamada previamente refund call_id=%s", call.CallID)
		return s.repo.FillMissingCallData(call)
	} else if status != "" {
		log.Printf("ℹ️ Llamada duplicada descartada call_id=%s con estado=%s", call.CallID, status)
		return nil
	}

	if err := s.repo.SaveIncomingCall(call); err != nil {
		return err
	}

		costResp, err := s.costClient.GetCallCost(call.CallID)
	if err != nil {
		// Si es error de cliente (4xx), marcamos como inválido
		var apiErr *client.CostAPIError
		if errors.As(err, &apiErr) && apiErr.IsClientError() {
			log.Printf("⚠️ Llamada inválida call_id=%s: %v", call.CallID, err)
			return s.repo.MarkCallAsInvalid(call.CallID)
		}

		log.Printf("⚠️ Error obteniendo costo para call_id=%s: %v", call.CallID, err)
		return s.repo.MarkCostAsFailed(call.CallID)
	}


	return s.repo.UpdateCallCost(call.CallID, costResp.Cost, costResp.Currency)
}
