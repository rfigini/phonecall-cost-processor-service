package application

import (
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
)

type MockCallRepository struct {
	Called     bool
	RefundData model.RefundCall
	ShouldErr  bool
}

// GetCallStatus implements repository.CallRepository.
func (m *MockCallRepository) GetCallStatus(callID string) (string, error) {
	panic("unimplemented")
}

func (m *MockCallRepository) MarkCostAsFailed(callID string) error {
	panic("unimplemented")
}

func (m *MockCallRepository) SaveIncomingCall(model.NewIncomingCall) error {
	return nil
}

func (m *MockCallRepository) UpdateCallCost(string, float64, string) error {
	return nil
}

func (m *MockCallRepository) ApplyRefund(refund model.RefundCall) error {
	m.Called = true
	m.RefundData = refund
	if m.ShouldErr {
		return errors.New("mock error")
	}
	return nil
}

func TestRefundCallUseCase_ApplyRefund(t *testing.T) {
	mockRepo := &MockCallRepository{}
	useCase := NewRefundCallUseCase(mockRepo)

	refund := model.RefundCall{
		CallID: "abc-123",
		Reason: "Test reason",
	}

	err := useCase.Execute(refund)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mockRepo.Called {
		t.Error("expected ApplyRefund to be called")
	}

	if mockRepo.RefundData != refund {
		t.Error("expected refund data to be passed correctly")
	}
}

func TestRefundCallUseCase_ApplyRefund_Error(t *testing.T) {
	mockRepo := &MockCallRepository{ShouldErr: true}
	useCase := NewRefundCallUseCase(mockRepo)

	refund := model.RefundCall{
		CallID: "abc-123",
		Reason: "Test reason",
	}

	err := useCase.Execute(refund)
	if err == nil {
		t.Error("expected error but got none")
	}
}
