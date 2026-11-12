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
	apiv2 "github.com/samantonio28/subscriber-inf/internal/api/v2"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	serviceredis "github.com/samantonio28/subscriber-inf/internal/service/redis"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App() {
	cfg, err := config.LoadConfig("configs/database.yaml")
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

	serviceRepo, err := service.NewServiceRepo(pool)
	if err != nil {
		log.Fatal("Failed to create service repo:", err)
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

	serverV2 := NewServerV2(serverImpl, serviceRepo)
	rV2 := apiv2.Handler(serverV2)

	rWithMiddleware := mux.NewRouter()
	rWithMiddleware.Use(CORSMiddleware())
	rWithMiddleware.Use(AuthMiddleware(userRepo))
	rWithMiddleware.Use(AccessLogMiddleware(logger))

	// /health — liveness: процесс жив (не проверяет зависимости).
	rWithMiddleware.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	// /ready — readiness: готов обслуживать запросы (Postgres + Redis доступны).
	rWithMiddleware.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","dependency":"postgres"}`))
			return
		}
		if err := redisClient.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","dependency":"redis"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}).Methods(http.MethodGet)

	rWithMiddleware.PathPrefix("/api/v2").Handler(rV2)
	rWithMiddleware.PathPrefix("/api/v1").Handler(http.StripPrefix("/api/v1", r))
	rWithMiddleware.PathPrefix("/").Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := http.Server{
		Addr:         ":" + port,
		Handler:      rWithMiddleware,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :" + port)
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}
