package delivery

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	serviceredis "github.com/samantonio28/subscriber-inf/internal/service/redis"
)

// BuildHandler собирает полный HTTP-обработчик приложения (репозитории →
// usecases → HTTP-хендлеры → middleware) из уже готовых соединений к
// PostgreSQL и Redis.
//
// Разделение на BuildHandler + App нужно, чтобы E2E-тест мог собрать тот же
// самый сервер, но с соединениями, указывающими на тестовый стенд
// (отдельные инстансы PostgreSQL/Redis из docker-compose.test.yml), а не на
// продакшн-конфиг.
func BuildHandler(pool *pgxpool.Pool, redisClient *redis.Client, lg *logger.LogrusLogger) (http.Handler, error) {
	repo, err := service.NewSubRepo(pool)
	if err != nil {
		return nil, err
	}

	promoRepo, err := service.NewPromocodeRepo(pool)
	if err != nil {
		return nil, err
	}

	planRepo, err := service.NewSubscriptionPlanRepo(pool)
	if err != nil {
		return nil, err
	}

	userRepo, err := service.NewUserRepo(pool)
	if err != nil {
		return nil, err
	}

	paymentRepo, err := service.NewPaymentRepo(pool)
	if err != nil {
		return nil, err
	}

	userServiceRepo, err := service.NewUserServiceRepo(pool)
	if err != nil {
		return nil, err
	}

	statsService, err := service.NewStatsService(pool, redisClient)
	if err != nil {
		return nil, err
	}

	planCache, err := serviceredis.NewSubscriptionPlanCache(redisClient)
	if err != nil {
		return nil, err
	}

	promoCache, err := serviceredis.NewPromocodeCache(redisClient)
	if err != nil {
		return nil, err
	}

	serverImpl, err := NewSubsServer(repo, promoRepo, planRepo, planCache, promoCache, statsService, userRepo, paymentRepo, userServiceRepo, lg)
	if err != nil {
		return nil, err
	}

	r := api.Handler(serverImpl)

	rWithMiddleware := mux.NewRouter()
	rWithMiddleware.Use(CORSMiddleware())
	rWithMiddleware.Use(AuthMiddleware(userRepo))
	rWithMiddleware.Use(AccessLogMiddleware(lg))
	rWithMiddleware.PathPrefix("/").Handler(r)

	return rWithMiddleware, nil
}
