package delivery

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/service/mysql"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func App(redisClient *redis.Client) {
	// Загружаем конфигурацию базы данных
	dbCfg, err := config.LoadDatabaseConfig("configs/database.yaml")
	if err != nil {
		log.Fatal("Failed to load database config:", err)
	}

	log.Printf("Using database type: %s", dbCfg.Type)

	var (
		repo         domain.SubscriptionRepository
		promoRepo    domain.PromocodeRepository
		planRepo     domain.SubscriptionPlanRepository
		userRepo     domain.UserRepository
		paymentRepo  domain.PaymentRepository
		statsService domain.StatsService
	)

	switch dbCfg.Type {
	case config.DBTypePostgres:
		pool, err := connectPostgres(dbCfg)
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL:", err)
		}
		defer pool.Close()

		repo, err = service.NewSubRepo(pool)
		if err != nil {
			log.Fatal("Failed to create sub repo:", err)
		}
		promoRepo, err = service.NewPromocodeRepo(pool)
		if err != nil {
			log.Fatal("Failed to create promocode repo:", err)
		}
		planRepo, err = service.NewSubscriptionPlanRepo(pool)
		if err != nil {
			log.Fatal("Failed to create subscription plan repo:", err)
		}
		userRepo, err = service.NewUserRepo(pool)
		if err != nil {
			log.Fatal("Failed to create user repo:", err)
		}
		paymentRepo, err = service.NewPaymentRepo(pool)
		if err != nil {
			log.Fatal("Failed to create payment repo:", err)
		}
		statsService, err = service.NewStatsService(pool, redisClient)
		if err != nil {
			log.Fatal("Failed to create stats repo:", err)
		}

	case config.DBTypeMySQL:
		db, err := connectMySQL(dbCfg)
		if err != nil {
			log.Fatal("Failed to connect to MySQL:", err)
		}
		defer db.Close()

		repo, err = mysql.NewSubRepo(db)
		if err != nil {
			log.Fatal("Failed to create sub repo (mysql):", err)
		}
		promoRepo, err = mysql.NewPromocodeRepo(db)
		if err != nil {
			log.Fatal("Failed to create promocode repo (mysql):", err)
		}
		planRepo, err = mysql.NewSubscriptionPlanRepo(db)
		if err != nil {
			log.Fatal("Failed to create subscription plan repo (mysql):", err)
		}
		userRepo, err = mysql.NewUserRepo(db)
		if err != nil {
			log.Fatal("Failed to create user repo (mysql):", err)
		}
		paymentRepo, err = mysql.NewPaymentRepo(db)
		if err != nil {
			log.Fatal("Failed to create payment repo (mysql):", err)
		}
		statsService, err = mysql.NewStatsService(db, redisClient)
		if err != nil {
			log.Fatal("Failed to create stats repo (mysql):", err)
		}

	default:
		log.Fatalf("Unsupported database type: %s", dbCfg.Type)
	}

	logger, err := logger.NewLogrusLogger("logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	serverImpl, err := NewSubsServer(repo, promoRepo, planRepo, statsService, userRepo, paymentRepo, logger)
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

func connectPostgres(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.Postgres == nil {
		return nil, fmt.Errorf("postgres config is missing")
	}
	poolConfig, err := cfg.Postgres.ToPgxPoolConfig()
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to PostgreSQL!")
	return pool, nil
}

func connectMySQL(cfg *config.DatabaseConfig) (*sql.DB, error) {
	if cfg.MySQL == nil {
		return nil, fmt.Errorf("mysql config is missing")
	}
	db, err := cfg.Open()
	if err != nil {
		return nil, err
	}
	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Successfully connected to MySQL!")
	return db, nil
}
