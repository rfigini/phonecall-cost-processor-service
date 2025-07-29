package application

import (
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/repository"
)

type IRefundCallUseCase interface {
	Execute(call model.RefundCall) error
}

type RefundCallUseCase struct {
	repo repository.CallRepository
}

func NewRefundCallUseCase(repo repository.CallRepository) *RefundCallUseCase {
	return &RefundCallUseCase{repo: repo}
}

func (uc *RefundCallUseCase) Execute(call model.RefundCall) error {
	return uc.repo.ApplyRefund(call)
}
