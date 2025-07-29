package service_test

import (
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/model/services"
	
)

// Mock del repositorio
type MockCallRepo struct {
	SavedCall       *model.NewIncomingCall
	UpdatedCallCost struct {
		CallID   string
		Cost     float64
		Currency string
	}
	SaveErr   error
	UpdateErr error
}

func (m *MockCallRepo) SaveIncomingCall(call model.NewIncomingCall) error {
	m.SavedCall = &call
	return m.SaveErr
}

func (m *MockCallRepo) ApplyRefund(model.RefundCall) error {
	return nil
}

func (m *MockCallRepo) UpdateCallCost(callID string, cost float64, currency string) error {
	m.UpdatedCallCost.CallID = callID
	m.UpdatedCallCost.Cost = cost
	m.UpdatedCallCost.Currency = currency
	return m.UpdateErr
}

// Mock del cliente
type MockCostClient struct {
	Response model.CostResponse
	Err      error
}

func (m *MockCostClient) GetCallCost(callID string) (model.CostResponse, error) {
	return m.Response, m.Err
}

func TestCallService_Process_Success(t *testing.T) {
	mockRepo := &MockCallRepo{}
	mockClient := &MockCostClient{
		Response: model.CostResponse{Cost: 5.0, Currency: "ARS"},
	}
	callService := service.NewCallService(mockRepo, mockClient)

	call := model.NewIncomingCall{
		CallID: "123",
	}

	err := callService.Process(call)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mockRepo.SavedCall == nil || mockRepo.SavedCall.CallID != "123" {
		t.Error("call was not saved properly")
	}

	if mockRepo.UpdatedCallCost.CallID != "123" || mockRepo.UpdatedCallCost.Cost != 5.0 || mockRepo.UpdatedCallCost.Currency != "ARS" {
		t.Error("call cost was not updated properly")
	}
}

func TestCallService_Process_CostClientFails(t *testing.T) {
	mockRepo := &MockCallRepo{}
	mockClient := &MockCostClient{Err: errors.New("API error")}
	callService := service.NewCallService(mockRepo, mockClient)

	call := model.NewIncomingCall{
		CallID: "456",
	}

	err := callService.Process(call)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mockRepo.UpdatedCallCost.Cost != 0 || mockRepo.UpdatedCallCost.Currency != "" {
		t.Error("cost update fallback not applied correctly")
	}
}
