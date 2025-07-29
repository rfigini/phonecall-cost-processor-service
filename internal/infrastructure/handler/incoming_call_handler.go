package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/rabbitmq/dto"
)

type IncomingCallHandler struct {
    useCase application.IIncomingCallUseCase
}

func NewIncomingCallHandler(useCase application.IIncomingCallUseCase) *IncomingCallHandler {
    return &IncomingCallHandler{useCase: useCase}
}

func (h *IncomingCallHandler) Handle(msg []byte) error {
    // 1) Unmarshal al DTO
    var d dto.NewIncomingCallDTO
    if err := json.Unmarshal(msg, &d); err != nil {
        log.Printf("❌ Error parseando DTO: %v\n", err)
        return err
    }

    // 2) Parseamos el timestamp y capturamos la variable
    startTime, err := time.Parse(time.RFC3339, d.StartTimestamp)
    if err != nil {
        return fmt.Errorf("start_timestamp inválido: %w", err)
    }

    // 3) Mapeo a dominio usando startTime
    call := model.NewIncomingCall{
        CallID:         d.CallID,
        Caller:         d.Caller,
        Receiver:       d.Receiver,
        DurationInSec:  d.DurationInSec,
        StartTimestamp: startTime.Format(time.RFC3339),
    }

    // 4) Delegamos al caso de uso
    if err := h.useCase.Execute(call); err != nil {
        log.Printf("❌ Error procesando llamada: %v\n", err)
        return err
    }

    log.Printf("📞 Llamada procesada: %+v\n", call)
    return nil
}
