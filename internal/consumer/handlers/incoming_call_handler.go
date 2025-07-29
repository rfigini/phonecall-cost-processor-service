package handlers

import (
	"encoding/json"
	"log"
	"phonecall-cost-processor-service/internal/client"
	"phonecall-cost-processor-service/internal/model"
	"phonecall-cost-processor-service/internal/repository"
)

func NewIncomingCallHandler(repo *repository.CallRepository, costClient *client.CostClient) func(body json.RawMessage) error {
	return func(body json.RawMessage) error {
		var call model.NewIncomingCall
		if err := json.Unmarshal(body, &call); err != nil {
			return err
		}

		costResp, err := costClient.GetCallCost(call.CallID)
		if err != nil {
			log.Printf("⚠️ Error obteniendo costo para %s: %v", call.CallID, err)
			call.CostFetchFailed = true
		} else {
			call.Cost = &costResp.Cost
			call.Currency = costResp.Currency
			call.CostFetchFailed = false
		}

		if err := repo.SaveIncomingCall(call); err != nil {
			return err
		}

		log.Printf("📞 Llamada procesada: %+v\n", call)
		return nil
	}
}
