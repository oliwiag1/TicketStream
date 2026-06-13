package main

import (
	"ticketstream/backend/internal/config"
	"ticketstream/backend/internal/worker/queue"
	"ticketstream/backend/pkg/broker"
	"ticketstream/backend/pkg/logger"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New(cfg.AppEnv)

	rabbitConn, rabbitChannel, err := broker.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		appLogger.Fatalf("rabbitmq connection failed: %v", err)
	}
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	consumer := queue.NewConsumer(rabbitChannel, appLogger)
	appLogger.Printf("worker started")

	if err := consumer.Start(); err != nil {
		appLogger.Fatalf("worker stopped with error: %v", err)
	}
}
