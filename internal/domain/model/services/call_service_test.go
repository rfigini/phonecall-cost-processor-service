package services

import (
	"errors"
	"testing"

	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/client"
)

// Mocks
// mockRepo implementa repository.CallRepository
type mockRepo struct {
	SaveErr    error
	SaveCalled bool
	SaveInput  model.NewIncomingCall

	MarkFailedErr    error
	MarkFailedCalled bool
	MarkFailedInput  string

	UpdateErr           error
	UpdateCalled        bool
	UpdateInputID       string
	UpdateInputCost     float64
	UpdateInputCur      string
	GetCallStatusOutput string
	GetCallStatusErr    error
	FillCalled          bool
	FillInput           model.NewIncomingCall
	FillFunc            func(call model.NewIncomingCall) error

	InvalidCalled bool
	InvalidInput  string
	InvalidFunc   func(callID string) error
}

func (m *mockRepo) FillMissingCallData(call model.NewIncomingCall) error {
	m.FillCalled = true
	m.FillInput = call
	if m.FillFunc != nil {
		return m.FillFunc(call)
	}
	return nil
}

func (m *mockRepo) MarkCallAsInvalid(callID string) error {
	m.InvalidCalled = true
	m.InvalidInput = callID
	if m.InvalidFunc != nil {
		return m.InvalidFunc(callID)
	}
	return nil
}

func (m *mockRepo) GetCallStatus(callID string) (string, error) {
	return m.GetCallStatusOutput, m.GetCallStatusErr
}

func (m *mockRepo) SaveIncomingCall(call model.NewIncomingCall) error {
	m.SaveCalled = true
	m.SaveInput = call
	return m.SaveErr
}

func (m *mockRepo) MarkCostAsFailed(callID string) error {
	m.MarkFailedCalled = true
	m.MarkFailedInput = callID
	return m.MarkFailedErr
}

func (m *mockRepo) UpdateCallCost(callID string, cost float64, currency string) error {
	m.UpdateCalled = true
	m.UpdateInputID = callID
	m.UpdateInputCost = cost
	m.UpdateInputCur = currency
	return m.UpdateErr
}

func (m *mockRepo) ApplyRefund(refund model.RefundCall) error {
	// not used in CallService
	return nil
}

type mockClient struct {
	GetErr      error
	Called      bool
	CalledInput string
	Resp        *model.CostResponse // ahora puntero
}

func (m *mockClient) GetCallCost(callID string) (*model.CostResponse, error) {
	m.Called = true
	m.CalledInput = callID
	return m.Resp, m.GetErr
}

// Tests
func TestProcess_SaveError(t *testing.T) {
	repo := &mockRepo{SaveErr: errors.New("save failed")}
	client := &mockClient{}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id1"}

	err := svc.Process(call)
	if err == nil || err.Error() != "save failed" {
		t.Fatalf("expected save error, got %v", err)
	}
	if !repo.SaveCalled {
		t.Error("SaveIncomingCall should be called")
	}
	if client.Called {
		t.Error("GetCallCost should not be called on save error")
	}
}

func TestProcess_CostError_MarkFailed(t *testing.T) {
	repo := &mockRepo{}
	client := &mockClient{GetErr: errors.New("client error")}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id2"}

	err := svc.Process(call)
	if err != nil {
		t.Fatalf("expected no error on mark failed, got %v", err)
	}
	if !repo.SaveCalled {
		t.Error("SaveIncomingCall should be called")
	}
	if !client.Called || client.CalledInput != "id2" {
		t.Error("GetCallCost should be called with correct ID")
	}
	if !repo.MarkFailedCalled || repo.MarkFailedInput != "id2" {
		t.Error("MarkCostAsFailed should be called with correct ID")
	}
}

func TestProcess_CostError_MarkFailedError(t *testing.T) {
	repo := &mockRepo{MarkFailedErr: errors.New("mark failed err")}
	client := &mockClient{GetErr: errors.New("client error")}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id3"}

	err := svc.Process(call)
	if err == nil || err.Error() != "mark failed err" {
		t.Fatalf("expected mark failed error, got %v", err)
	}
	if !repo.MarkFailedCalled {
		t.Error("MarkCostAsFailed should be called")
	}
}

func TestProcess_Success(t *testing.T) {
	repo := &mockRepo{}
	client := &mockClient{Resp: &model.CostResponse{Cost: 9.99, Currency: "USD"}}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id4"}

	err := svc.Process(call)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if !repo.SaveCalled {
		t.Error("SaveIncomingCall should be called")
	}
	if !client.Called {
		t.Error("GetCallCost should be called")
	}
	if !repo.UpdateCalled {
		t.Error("UpdateCallCost should be called")
	}
	if repo.UpdateInputID != "id4" || repo.UpdateInputCost != 9.99 || repo.UpdateInputCur != "USD" {
		t.Errorf("UpdateCallCost called with wrong args: %v, %v, %v", repo.UpdateInputID, repo.UpdateInputCost, repo.UpdateInputCur)
	}
}

func TestProcess_UpdateError(t *testing.T) {
	repo := &mockRepo{UpdateErr: errors.New("update error")}
	client := &mockClient{Resp: &model.CostResponse{Cost: 1.23, Currency: "EUR"}}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id5"}

	err := svc.Process(call)
	if err == nil || err.Error() != "update error" {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestProcess_RefundedCall_SkipsProcessing(t *testing.T) {
	repo := &mockRepo{
		GetCallStatusOutput: "REFUNDED",
	}
	client := &mockClient{}
	svc := NewCallService(repo, client)

	call := model.NewIncomingCall{CallID: "id_refunded"}
	err := svc.Process(call)

	if err != nil {
		t.Fatalf("expected no error for refunded call, got %v", err)
	}
	if repo.SaveCalled {
		t.Error("SaveIncomingCall should not be called for refunded call")
	}
	if client.Called {
		t.Error("GetCallCost should not be called for refunded call")
	}
}

func TestProcess_DuplicatedCall_Discarded(t *testing.T) {
	repo := &mockRepo{
		GetCallStatusOutput: "PROCESSED",
	}
	client := &mockClient{}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id_duplicate"}

	err := svc.Process(call)
	if err != nil {
		t.Fatalf("expected no error for duplicated call, got %v", err)
	}
	if repo.SaveCalled {
		t.Error("SaveIncomingCall should not be called for duplicated call")
	}
	if client.Called {
		t.Error("GetCallCost should not be called for duplicated call")
	}
}

func TestProcess_CostError_Client4xx_MarkInvalid(t *testing.T) {
	repo := &mockRepo{}
	apiErr := &client.CostAPIError{StatusCode: 404}
	client := &mockClient{GetErr: apiErr}
	svc := NewCallService(repo, client)
	call := model.NewIncomingCall{CallID: "id_invalid"}

	err := svc.Process(call)
	if err != nil {
		t.Fatalf("expected no error on mark invalid, got %v", err)
	}
}

func TestProcess_RefundedCall_FillDataFails(t *testing.T) {
	repo := &mockRepo{
		GetCallStatusOutput: "REFUNDED",
	}
	repo.FillFunc = func(call model.NewIncomingCall) error {
		return errors.New("fill failed")
	}
	client := &mockClient{}
	svc := NewCallService(repo, client)

	call := model.NewIncomingCall{CallID: "id_refund_fill_fail"}
	err := svc.Process(call)

	if err == nil || err.Error() != "fill failed" {
		t.Fatalf("expected fill error, got %v", err)
	}
	if !repo.FillCalled {
		t.Error("FillMissingCallData should be called for refunded call")
	}
}

func TestProcess_CostError_Client4xx_MarkInvalidFails(t *testing.T) {
	repo := &mockRepo{}
	repo.InvalidFunc = func(callID string) error {
		return errors.New("invalid mark failed")
	}
	apiErr := &client.CostAPIError{StatusCode: 400}
	client := &mockClient{GetErr: apiErr}
	svc := NewCallService(repo, client)

	call := model.NewIncomingCall{CallID: "id_invalid_fail"}
	err := svc.Process(call)

	if err == nil || err.Error() != "invalid mark failed" {
		t.Fatalf("expected invalid mark error, got %v", err)
	}
	if !repo.InvalidCalled || repo.InvalidInput != "id_invalid_fail" {
		t.Error("MarkCallAsInvalid should be called with correct call ID")
	}
}
