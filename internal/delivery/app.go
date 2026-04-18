package delivery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App(redisClient *redis.Client) {
	cfg, err := config.LoadConfig("configs/postgres.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

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

	repo, err := service.NewSubRepo(pool)
	if err != nil {
		log.Fatal("Failed to create sub repo:", err)
	}

	promoRepo, err := service.NewPromocodeRepo(pool)
	if err != nil {
		log.Fatal("Failed to create promocode repo:", err)
	}

	planRepo, err := service.NewSubscriptionPlanRepo(pool)
	if err != nil {
		log.Fatal("Failed to create subscription plan repo:", err)
	}

	statsService, err := service.NewStatsService(pool, redisClient)
	if err != nil {
		log.Fatal("Failed to create stats repo:", err)
	}

	logger, err := logger.NewLogrusLogger("logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	serverImpl, err := NewSubsServer(repo, promoRepo, planRepo, statsService, logger)
	if err != nil {
		log.Fatal("Failed to create server implementation:", err)
	}

	r := api.Handler(serverImpl)

	rWithMiddleware := mux.NewRouter()
	rWithMiddleware.Use(CORSMiddleware())
	rWithMiddleware.Use(AccessLogMiddleware(logger))
	rWithMiddleware.PathPrefix("/").Handler(r)

	server := http.Server{
		Addr:         ":8080",
		Handler:      rWithMiddleware,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}
