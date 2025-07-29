package application

import (
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/model/services"
)


type IIncomingCallUseCase interface {
	Execute(call model.NewIncomingCall) error
}

type IncomingCallUseCase struct {
	callService service.ICallService
}

func NewIncomingCallUseCase(callService service.ICallService) *IncomingCallUseCase {
	return &IncomingCallUseCase{callService}
}

func (uc *IncomingCallUseCase) Execute(call model.NewIncomingCall) error {
	return uc.callService.Process(call)
}



