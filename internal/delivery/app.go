package delivery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App() {
	// Путь к конфигу можно переопределить через CONFIG_PATH — например, чтобы
	// запустить приложение против тестового стенда (configs/postgres_test.yaml).
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/postgres.yaml"
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = cfg.Redis.Addr()
	}
	redisClient, err := redis.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	log.Println("Successfully connected to Redis!")

	poolConfig, err := cfg.Postgres.ToPgxPoolConfig()
	if err != nil {
		log.Fatal("Failed to create pool config:", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer pool.Close()

	log.Println("Successfully connected to PostgreSQL!")

	lg, err := logger.NewLogrusLogger("logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	handler, err := BuildHandler(pool, redisClient, lg)
	if err != nil {
		log.Fatal("Failed to build handler:", err)
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}
