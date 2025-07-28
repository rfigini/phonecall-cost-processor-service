package main

import (
	"fmt"
	"log"

	"phonecall-cost-processor-service/internal/config"
	"phonecall-cost-processor-service/internal/infrastructure"
	"phonecall-cost-processor-service/internal/consumer"
	"phonecall-cost-processor-service/internal/repository"

)

func main() {
	cfg := config.Load()

	fmt.Println("📦 Configuración cargada:")
	fmt.Println("RabbitMQ URL:", cfg.RabbitURL)
	fmt.Println("DB URL:", cfg.DBUrl)

	// Conexión a PostgreSQL
	db, err := infrastructure.NewPostgresConnection(cfg.DBUrl)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ Conexión a PostgreSQL exitosa")

	// Conexión a RabbitMQ
	rabbitConn, rabbitCh, err := infrastructure.NewRabbitConn(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	defer rabbitCh.Close()
	fmt.Println("✅ Conexión a RabbitMQ exitosa")
	callRepository := repository.NewCallRepository(db)

	err = consumer.StartConsumingMessages(rabbitCh, cfg.RabbitQueue, callRepository) 

	if err != nil {
		log.Fatalf("❌ Error iniciando consumidor: %v", err)
	}

	// Mantener el programa vivo

	select {}
}
