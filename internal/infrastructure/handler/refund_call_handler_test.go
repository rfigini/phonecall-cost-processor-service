package handler_test

import (
	"encoding/json"
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/infrastructure/handler"
)

// Mock del RefundCallUseCase
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

	refund := model.RefundCall{
		CallID: "abc-123",
		Reason: "Test reason",
	}
	data, _ := json.Marshal(refund)

	err := h.Handle(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mockUC.Called {
		t.Error("expected ApplyRefund to be called")
	}
	if mockUC.Input != refund {
		t.Errorf("expected input %+v but got %+v", refund, mockUC.Input)
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
		t.Error("ApplyRefund should not be called on invalid JSON")
	}
}

func TestRefundCallHandler_Handle_UseCaseError(t *testing.T) {
	mockUC := &MockRefundCallUseCase{ShouldFail: true}
	h := handler.NewRefundCallHandler(mockUC)

	refund := model.RefundCall{
		CallID: "abc-123",
		Reason: "fail reason",
	}
	data, _ := json.Marshal(refund)

	err := h.Handle(data)
	if err == nil {
		t.Error("expected error from ApplyRefund")
	}
}
