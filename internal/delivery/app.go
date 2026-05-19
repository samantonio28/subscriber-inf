package delivery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	serviceredis "github.com/samantonio28/subscriber-inf/internal/service/redis"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App() {
	cfg, err := config.LoadConfig("configs/postgres.yaml")
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

	userRepo, err := service.NewUserRepo(pool)
	if err != nil {
		log.Fatal("Failed to create user repo:", err)
	}

	paymentRepo, err := service.NewPaymentRepo(pool)
	if err != nil {
		log.Fatal("Failed to create payment repo:", err)
	}

	userServiceRepo, err := service.NewUserServiceRepo(pool)
	if err != nil {
		log.Fatal("Failed to create user service repo:", err)
	}

	statsService, err := service.NewStatsService(pool, redisClient)
	if err != nil {
		log.Fatal("Failed to create stats repo:", err)
	}

	planCache, err := serviceredis.NewSubscriptionPlanCache(redisClient)
	if err != nil {
		log.Fatal("Failed to create subscription plan cache:", err)
	}

	promoCache, err := serviceredis.NewPromocodeCache(redisClient)
	if err != nil {
		log.Fatal("Failed to create promocode cache:", err)
	}

	logger, err := logger.NewLogrusLogger("logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	serverImpl, err := NewSubsServer(repo, promoRepo, planRepo, planCache, promoCache, statsService, userRepo, paymentRepo, userServiceRepo, logger)
	if err != nil {
		log.Fatal("Failed to create server implementation:", err)
	}

	r := api.Handler(serverImpl)

	rWithMiddleware := mux.NewRouter()
	rWithMiddleware.Use(CORSMiddleware())
	rWithMiddleware.Use(AuthMiddleware(userRepo))
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
