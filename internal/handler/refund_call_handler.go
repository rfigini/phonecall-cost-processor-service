package handler

import (
	"encoding/json"
	"log"
	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/domain/model"
)

type RefundCallHandler struct {
	useCase *application.RefundCallUseCase
}

func NewRefundCallHandler(useCase *application.RefundCallUseCase) *RefundCallHandler {
	return &RefundCallHandler{useCase: useCase}
}

func (h *RefundCallHandler) Handle(msg []byte) error {
	var refund model.RefundCall
	if err := json.Unmarshal(msg, &refund); err != nil {
		log.Printf("❌ Error parseando refund: %v\n", err)
		return err
	}

	err := h.useCase.ApplyRefund(refund)
	if err != nil {
		log.Printf("❌ Error aplicando refund: %v\n", err)
		return err
	}

	log.Printf("💸 Refund aplicado correctamente: %+v\n", refund)
	return nil
}
