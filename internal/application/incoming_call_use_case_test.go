package application

import (
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
)

type MockCallService struct {
	Called   bool
	CallData model.NewIncomingCall
	ShouldErr bool
}

func (m *MockCallService) Process(call model.NewIncomingCall) error {
	m.Called = true
	m.CallData = call
	if m.ShouldErr {
		return errors.New("mock error")
	}
	return nil
}

func TestIncomingCallUseCase_Execute(t *testing.T) {
	mockService := &MockCallService{}
	useCase := NewIncomingCallUseCase(mockService)

	call := model.NewIncomingCall{
		CallID:         "123",
		Caller:         "+123456",
		Receiver:       "+654321",
		DurationInSec:  60,
		StartTimestamp: "2025-07-25T03:00:00Z",
	}

	err := useCase.Execute(call)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mockService.Called {
		t.Error("expected Process to be called")
	}

	if mockService.CallData != call {
		t.Errorf("expected call data to be passed correctly")
	}
}

func TestIncomingCallUseCase_Execute_Error(t *testing.T) {
	mockService := &MockCallService{ShouldErr: true}
	useCase := NewIncomingCallUseCase(mockService)

	call := model.NewIncomingCall{CallID: "error-case"}

	err := useCase.Execute(call)

	if err == nil {
		t.Error("expected an error but got nil")
	}
	if !mockService.Called {
		t.Error("expected Process to be called")
	}
}
