package handler_test

import (
	"encoding/json"
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/handler"
)

// Mock del use case
type MockIncomingCallUseCase struct {
	Called   bool
	Input    model.NewIncomingCall
	ShouldErr bool
}

func (m *MockIncomingCallUseCase) Execute(call model.NewIncomingCall) error {
	m.Called = true
	m.Input = call
	if m.ShouldErr {
		return errors.New("use case error")
	}
	return nil
}

func TestIncomingCallHandler_Handle_Success(t *testing.T) {
	mockUC := &MockIncomingCallUseCase{}
	h := handler.NewIncomingCallHandler(mockUC)

	call := model.NewIncomingCall{
		CallID:         "123",
		Caller:         "+123",
		Receiver:       "+456",
		DurationInSec:  60,
		StartTimestamp: "2025-07-25T03:00:00Z",
	}
	jsonBytes, _ := json.Marshal(call)

	err := h.Handle(jsonBytes)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !mockUC.Called {
		t.Error("expected Execute to be called")
	}
	if mockUC.Input != call {
		t.Errorf("expected input %+v but got %+v", call, mockUC.Input)
	}
}

func TestIncomingCallHandler_Handle_InvalidJSON(t *testing.T) {
	mockUC := &MockIncomingCallUseCase{}
	h := handler.NewIncomingCallHandler(mockUC)

	err := h.Handle([]byte("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if mockUC.Called {
		t.Error("Execute should not be called")
	}
}

func TestIncomingCallHandler_Handle_UseCaseError(t *testing.T) {
	mockUC := &MockIncomingCallUseCase{ShouldErr: true}
	h := handler.NewIncomingCallHandler(mockUC)

	call := model.NewIncomingCall{
		CallID:         "123",
		Caller:         "+123",
		Receiver:       "+456",
		DurationInSec:  60,
		StartTimestamp: "2025-07-25T03:00:00Z",
	}
	jsonBytes, _ := json.Marshal(call)

	err := h.Handle(jsonBytes)
	if err == nil {
		t.Error("expected error from use case")
	}
}
