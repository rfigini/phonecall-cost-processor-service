package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/rabbitmq/dto"
)

// RefundCallHandler procesa mensajes de tipo "refund_call"
type RefundCallHandler struct {
	useCase application.IRefundCallUseCase
}

func NewRefundCallHandler(useCase application.IRefundCallUseCase) *RefundCallHandler {
	return &RefundCallHandler{useCase: useCase}
}

// Handle recibe el JSON del mensaje, deserializa al DTO, mapea al dominio y ejecuta el caso de uso
func (h *RefundCallHandler) Handle(msg []byte) error {
	// 1) Deserializar al DTO de infraestructura
	var d dto.RefundCallDTO
	if err := json.Unmarshal(msg, &d); err != nil {
		log.Printf("❌ Error parseando DTO de refund: %v", err)
		return fmt.Errorf("payload inválido para refund_call: %w", err)
	}

	// 2) Mapear a modelo de dominio
	refund := model.RefundCall{
		CallID: d.CallID,
		Reason: d.Reason,
	}

	// 3) Ejecutar caso de uso
	if err := h.useCase.Execute(refund); err != nil {
		log.Printf("❌ Error aplicando refund: %v", err)
		return err
	}

	log.Printf("💸 Refund aplicado correctamente: %+v", refund)
	return nil
}