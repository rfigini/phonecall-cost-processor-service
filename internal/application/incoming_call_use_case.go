package application

import (
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/model/service"
)

type IncomingCallUseCase struct {
	callService *service.CallService
}

func NewIncomingCallUseCase(callService *service.CallService) *IncomingCallUseCase {
	return &IncomingCallUseCase{callService}
}

func (uc *IncomingCallUseCase) Execute(call model.NewIncomingCall) error {
	return uc.callService.Process(call)
}
