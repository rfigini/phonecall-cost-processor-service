package handler

import (
	"encoding/json"
	"log"
	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/domain/model"
)

type IncomingCallHandler struct {
	useCase *application.IncomingCallUseCase
}

func NewIncomingCallHandler(useCase *application.IncomingCallUseCase) *IncomingCallHandler {
	return &IncomingCallHandler{useCase: useCase}
}

func (h *IncomingCallHandler) Handle(msg []byte) error {
	var call model.NewIncomingCall
	if err := json.Unmarshal(msg, &call); err != nil {
		log.Printf("❌ Error parseando llamada: %v\n", err)
		return err
	}

	err := h.useCase.Execute(call)
	if err != nil {
		log.Printf("❌ Error procesando llamada: %v\n", err)
		return err
	}

	log.Printf("📞 Llamada procesada: %+v\n", call)
	return nil
}
