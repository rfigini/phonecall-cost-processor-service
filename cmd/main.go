package main

import (
	"fmt"
	"phonecall-cost-processor-service/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Println("📦 Configuración cargada:")
	fmt.Println("RabbitMQ URL:", cfg.RabbitURL)
	fmt.Println("RabbitMQ Queue:", cfg.RabbitQueue)
	fmt.Println("DB URL:", cfg.DBUrl)
	fmt.Println("Cost API URL:", cfg.CostAPIUrl)
}
