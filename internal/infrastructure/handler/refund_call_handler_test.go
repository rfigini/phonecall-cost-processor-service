package handler_test

import (
	"encoding/json"
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/handler"
	"phonecall-cost-processor-service/internal/infrastructure/rabbitmq/dto"
)


type MockRefundCallUseCase struct {
	Called     bool
	Input      model.RefundCall
	ShouldFail bool
}

func (m *MockRefundCallUseCase) Execute(refund model.RefundCall) error {
	m.Called = true
	m.Input = refund
	if m.ShouldFail {
		return errors.New("apply refund failed")
	}
	return nil
}

func TestRefundCallHandler_Handle_Success(t *testing.T) {
	mockUC := &MockRefundCallUseCase{}
	h := handler.NewRefundCallHandler(mockUC)

	d := dto.RefundCallDTO{
		CallID: "550e8400-e29b-41d4-a716-446655440000",
		Reason: "Test reason",
	}
	msg, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("error marshaling DTO: %v", err)
	}

	err = h.Handle(msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !mockUC.Called {
		t.Error("expected Execute to be called")
	}

	expected := model.RefundCall{CallID: d.CallID, Reason: d.Reason}
	if mockUC.Input != expected {
		t.Errorf("expected input %+v but got %+v", expected, mockUC.Input)
	}
}

func TestRefundCallHandler_Handle_InvalidJSON(t *testing.T) {
	mockUC := &MockRefundCallUseCase{}
	h := handler.NewRefundCallHandler(mockUC)

	err := h.Handle([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if mockUC.Called {
		t.Error("Execute should not be called on invalid JSON")
	}
}

func TestRefundCallHandler_Handle_UseCaseError(t *testing.T) {
	mockUC := &MockRefundCallUseCase{ShouldFail: true}
	h := handler.NewRefundCallHandler(mockUC)

	d := dto.RefundCallDTO{
		CallID: "550e8400-e29b-41d4-a716-446655440000",
		Reason: "fail reason",
	}
	msg, _ := json.Marshal(d)

	err := h.Handle(msg)
	if err == nil {
		t.Error("expected error from use case")
	}
}
