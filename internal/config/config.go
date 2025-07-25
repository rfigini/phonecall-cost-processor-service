package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitURL   string
	RabbitQueue string
	DBUrl       string
	CostAPIUrl  string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, trying system environment variables")
	}

	return Config{
		RabbitURL:   os.Getenv("RABBITMQ_URL"),
		RabbitQueue: os.Getenv("RABBITMQ_QUEUE"),
		DBUrl:       os.Getenv("DB_URL"),
		CostAPIUrl:  os.Getenv("COST_API_URL"),
	}
}
