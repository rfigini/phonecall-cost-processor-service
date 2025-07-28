package consumer

import (
	"encoding/json"
	"log"

	"phonecall-cost-processor-service/internal/model"
	"phonecall-cost-processor-service/internal/repository"

	"github.com/streadway/amqp"
)

func StartConsumingMessages(ch *amqp.Channel, queueName string, callRepo *repository.CallRepository) error {
	_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("📡 Esperando mensajes...")

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

			switch msgType {
			case "new_incoming_call":
				var call model.NewIncomingCall
				if err := json.Unmarshal(raw["body"], &call); err != nil {
					log.Printf("❌ Error parseando llamada: %v\n", err)
					continue
				}

				if err := callRepo.SaveIncomingCall(call); err != nil {
					log.Printf("❌ Error guardando llamada: %v\n", err)
					continue
				}

				log.Printf("📞 Nueva llamada: %+v\n", call)

			case "refund_call":
				var refund model.RefundCall
				if err := json.Unmarshal(raw["body"], &refund); err != nil {
					log.Printf("❌ Error parseando refund: %v\n", err)
					continue
				}

				if err := callRepo.ApplyRefund(refund); err != nil {
					log.Printf("❌ Error aplicando refund: %v\n", err)
					continue
				}

				log.Printf("💸 Devolución recibida: %+v\n", refund)

			default:
				log.Printf("⚠️ Tipo desconocido: %s\n", msgType)
			}
		}
	}()

	return nil
}