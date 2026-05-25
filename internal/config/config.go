package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQURL string
	QueueName   string
	Port        string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
		QueueName:   os.Getenv("RABBITMQ_QUEUE_NAME_CONSUMING"),
		Port:        os.Getenv("SERVER_PORT"),
	}
}