package consumer

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

// Handler representa una interfaz polimórfica para procesar distintos tipos de mensajes.
type Handler interface {
	Handle([]byte) error
}

func StartConsumingMessages(ch *amqp.Channel, queueName string, handlers map[string]Handler) error {
	_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return err
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

			handler, ok := handlers[msgType]
			if !ok {
				log.Printf("⚠️ Tipo de mensaje desconocido: %s\n", msgType)
				continue
			}

			if err := handler.Handle(raw["body"]); err != nil {
				log.Printf("❌ Error procesando mensaje tipo %s: %v\n", msgType, err)
			}
		}
	}()

	return nil
}
