package repository

import "phonecall-cost-processor-service/internal/domain/model"

type CallRepository interface {
	SaveIncomingCall(model.NewIncomingCall) error
	UpdateCallCost(callID string, cost float64, currency string) error
	ApplyRefund(model.RefundCall) error
}
