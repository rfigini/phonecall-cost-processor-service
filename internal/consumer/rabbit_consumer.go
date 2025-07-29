package consumer

import (
	"encoding/json"
	"log"
	"phonecall-cost-processor-service/internal/client"
	"phonecall-cost-processor-service/internal/consumer/handlers"
	"phonecall-cost-processor-service/internal/repository"

	"github.com/streadway/amqp"
)

func StartConsumingMessages(ch *amqp.Channel, queueName string, callRepo *repository.CallRepository, costClient *client.CostClient) error {
	_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// 🎯 Mapa extensible de handlers
	handlerMap := map[string]HandlerFunc{
		"new_incoming_call": handlers.NewIncomingCallHandler(callRepo, costClient),
		"refund_call":       handlers.NewRefundCallHandler(callRepo),
	}

	go func() {
		for msg := range msgs {
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(msg.Body, &raw); err != nil {
				log.Printf("❌ Error parseando mensaje: %v\n", err)
				continue
			}

			var msgType string
			if err := json.Unmarshal(raw["type"], &msgType); err != nil {
				log.Printf("❌ Error leyendo tipo: %v\n", err)
				continue
			}

			handler, ok := handlerMap[msgType]
			if !ok {
				log.Printf("⚠️ Tipo de mensaje desconocido: %s\n", msgType)
				continue
			}

			if err := handler(raw["body"]); err != nil {
				log.Printf("❌ Error procesando mensaje tipo %s: %v\n", msgType, err)
			}
		}
	}()

	return nil
}
