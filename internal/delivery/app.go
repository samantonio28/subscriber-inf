package delivery

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
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
}
