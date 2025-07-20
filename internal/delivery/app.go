package delivery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
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

	r := mux.NewRouter()
	// r.Use(AccessLogMiddleware(logger))

	handler, err := NewSubsHandler(repo)
	if err != nil {
		log.Fatal("Failed to create sub hander:", err)
	}

	r.HandleFunc("/subscriptions", handler.CreateSubscription).Methods("POST")
	// r.HandleFunc("/subscriptions", ).Methods("")
	// r.HandleFunc("/subscriptions/{id}", ).Methods("")
	// r.HandleFunc("/subscriptions/{id}", ).Methods("")
	// r.HandleFunc("/subscriptions/{id}", ).Methods("")
	// r.HandleFunc("/total_costs", ).Methods("")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8000"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	han := corsMiddleware.Handler(r)

	server := http.Server{
		Addr:         ":8080",
		Handler:      han,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}
