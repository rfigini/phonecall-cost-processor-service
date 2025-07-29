package handlers

import (
	"encoding/json"
	"log"
	"phonecall-cost-processor-service/internal/model"
	"phonecall-cost-processor-service/internal/repository"
)

func NewRefundCallHandler(repo *repository.CallRepository) func(body json.RawMessage) error {
	return func(body json.RawMessage) error {
		var refund model.RefundCall
		if err := json.Unmarshal(body, &refund); err != nil {
			return err
		}
		if err := repo.ApplyRefund(refund); err != nil {
			return err
		}
		log.Printf("💸 Devolución aplicada: %+v\n", refund)
		return nil
	}
}
