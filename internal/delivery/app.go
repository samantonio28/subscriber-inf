package delivery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App() {
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

	logger, err := logger.NewLogrusLogger("logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	r := mux.NewRouter()
	r.Use(AccessLogMiddleware(logger))

	handler, err := NewSubsHandler(repo, logger)
	if err != nil {
		log.Fatal("Failed to create sub hander:", err)
	}

	r.Use(AccessLogMiddleware(logger))
	r.HandleFunc("/subscriptions", handler.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", handler.GetSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", handler.DeleteSubscription).Methods("DELETE")
	r.HandleFunc("/subscriptions/{id}", handler.GetSubscription).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", handler.UpdateSubscription).Methods("PUT")
	r.HandleFunc("/total_costs", handler.GetTotalCosts).Methods("GET")

	server := http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}
