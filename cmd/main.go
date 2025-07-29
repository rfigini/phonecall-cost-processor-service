package main

import (
	"log"
	"phonecall-cost-processor-service/internal/application"
	"phonecall-cost-processor-service/internal/config"
	"phonecall-cost-processor-service/internal/consumer"
	"phonecall-cost-processor-service/internal/domain/service"
	"phonecall-cost-processor-service/internal/handler"
	"phonecall-cost-processor-service/internal/infrastructure"
	"phonecall-cost-processor-service/internal/infrastructure/client"
	"phonecall-cost-processor-service/internal/infrastructure/repository"
)

func main() {
	cfg := config.Load()

	// PostgreSQL
	db, err := infrastructure.NewPostgresConnection(cfg.DBUrl)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()

	// RabbitMQ
	rabbitConn, rabbitCh, err := infrastructure.NewRabbitConn(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	defer rabbitCh.Close()

	// Dependencias
	callRepo := repository.NewPostgresCallRepository(db)
	costClient := client.NewHTTPCostClient(cfg.CostAPIUrl)
	callService := service.NewCallService(callRepo, costClient)


	// Casos de uso
	incomingUseCase := application.NewIncomingCallUseCase(callService)
	refundUseCase := application.NewRefundCallUseCase(callRepo)

	// Handlers
	incomingHandler := handler.NewIncomingCallHandler(incomingUseCase)
	refundHandler := handler.NewRefundCallHandler(refundUseCase)

	// Map de handlers
	handlerMap := map[string]consumer.HandlerFunc{
		"new_incoming_call": incomingHandler.Handle,
		"refund_call":       refundHandler.Handle,
	}

	// Consumidor
	if err := consumer.StartConsumingMessages(rabbitCh, cfg.RabbitQueue, handlerMap); err != nil {
		log.Fatalf("❌ Error iniciando consumidor: %v", err)
	}

	select {}
}
