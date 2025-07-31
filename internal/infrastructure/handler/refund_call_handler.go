package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/rabbitmq/dto"
)

type RefundCallHandler struct {
	useCase application.IRefundCallUseCase
}

func NewRefundCallHandler(useCase application.IRefundCallUseCase) *RefundCallHandler {
	return &RefundCallHandler{useCase: useCase}
}

func (h *RefundCallHandler) Handle(msg []byte) error {
	var d dto.RefundCallDTO
	if err := json.Unmarshal(msg, &d); err != nil {
		log.Printf("❌ Error parseando DTO de refund: %v", err)
		return fmt.Errorf("payload inválido para refund_call: %w", err)
	}

	refund := model.RefundCall{
		CallID: d.CallID,
		Reason: d.Reason,
	}

	if err := h.useCase.Execute(refund); err != nil {
		log.Printf("❌ Error aplicando refund: %v", err)
		return err
	}

	log.Printf("💸 Refund aplicado correctamente: %+v", refund)
	return nil
}