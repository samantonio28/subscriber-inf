package main

import (
	"log"
	"os"

	"github.com/samantonio28/subscriber-inf/internal/delivery"
	"github.com/samantonio28/subscriber-inf/internal/redis"
)

func main() {
	// Инициализация Redis
	redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	log.Println("Successfully connected to Redis!")

	// Запуск приложения
	delivery.App(redisClient)
}
