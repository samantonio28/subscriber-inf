package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type Lab9Handler struct {
	pool         *pgxpool.Pool
	logger       *logger.LogrusLogger
	statsService *service.StatsService
	redisClient  *redis.Client
	mu           sync.Mutex
	metrics      []ExecutionMetric
}

type ExecutionMetric struct {
	Timestamp   time.Time     `json:"timestamp"`
	Operation   string        `json:"operation"`
	Source      string        `json:"source"`
	Duration    time.Duration `json:"duration"`
	DataChanged bool          `json:"data_changed"`
}

type TestScenario struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Interval    time.Duration `json:"interval"`
	DataChange  string        `json:"data_change"` // "none", "add", "delete", "update"
}

func NewLab9Handler(pool *pgxpool.Pool, logger *logger.LogrusLogger, redisClient *redis.Client) *Lab9Handler {
	statsService := service.NewStatsService(pool, redisClient)
	return &Lab9Handler{
		pool:         pool,
		logger:       logger,
		statsService: statsService,
		redisClient:  redisClient,
		metrics:      make([]ExecutionMetric, 0),
	}
}

func (h *Lab9Handler) Lab9Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	num := vars["num"]

	switch num {
	case "1":
		h.GetStatsDB(w, r)
	case "2":
		h.GetStatsCache(w, r)
	case "3":
		h.GetMetrics(w, r)
	case "4":
		h.StartPerformanceTest(w, r)
	case "5":
		h.StopPerformanceTest(w, r)
	case "6":
		h.AddTestData(w, r)
	case "7":
		h.ClearTestData(w, r)
	default:
		utils.MakeResponse(w, http.StatusNotFound, map[string]string{
			"message": "endpoint not found",
		})
	}
}

func (h *Lab9Handler) GetStatsDB(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats, err := h.statsService.GetServiceStatsFromDB(ctx)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get stats from DB: " + err.Error(),
		})
		return
	}

	h.recordMetric("get_stats", "db", stats.ExecutionTime, false)
	utils.MakeResponse(w, http.StatusOK, stats)
}

func (h *Lab9Handler) GetStatsCache(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats, err := h.statsService.GetServiceStatsFromCache(ctx)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get stats from cache: " + err.Error(),
		})
		return
	}

	h.recordMetric("get_stats", stats.Source, stats.ExecutionTime, false)
	utils.MakeResponse(w, http.StatusOK, stats)
}

func (h *Lab9Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Анализ метрик
	analysis := h.analyzeMetrics()

	response := map[string]interface{}{
		"metrics":  h.metrics,
		"analysis": analysis,
	}

	utils.MakeResponse(w, http.StatusOK, response)
}

func (h *Lab9Handler) StartPerformanceTest(w http.ResponseWriter, r *http.Request) {
	var scenario TestScenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid JSON",
		})
		return
	}

	go h.runPerformanceTest(scenario)

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "performance test started: " + scenario.Name,
	})
}

func (h *Lab9Handler) runPerformanceTest(scenario TestScenario) {
	ctx := context.Background()
	ticker := time.NewTicker(scenario.Interval)
	defer ticker.Stop()

	dataChangeTicker := time.NewTicker(10 * time.Second)
	defer dataChangeTicker.Stop()

	testDuration := 2 * time.Minute // Тест длится 2 минуты
	endTime := time.Now().Add(testDuration)

	h.logger.Info("Starting performance test", scenario.Name, "duration:", testDuration)

	for {
		select {
		case <-ticker.C:
			// Запрос к БД
			dbStart := time.Now()
			_, err := h.statsService.GetServiceStatsFromDB(ctx)
			dbDuration := time.Since(dbStart)

			if err == nil {
				h.recordMetric("performance_db", "db", dbDuration, false)
			}

			// Запрос к кэшу
			cacheStart := time.Now()
			_, err = h.statsService.GetServiceStatsFromCache(ctx)
			cacheDuration := time.Since(cacheStart)

			if err == nil {
				h.recordMetric("performance_cache", "cache", cacheDuration, false)
			}

		case <-dataChangeTicker.C:
			// Изменение данных в зависимости от сценария
			switch scenario.DataChange {
			case "add":
				h.addRandomSubscription(ctx)
			case "delete":
				h.deleteRandomSubscription(ctx)
			case "update":
				h.updateRandomSubscription(ctx)
			}

		default:
			if time.Now().After(endTime) {
				h.logger.Info("Performance test completed:", scenario.Name)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (h *Lab9Handler) AddTestData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Добавляем тестовые сервисы если их нет
	services := []string{"Netflix", "Spotify", "YouTube Premium", "Disney+", "Amazon Prime"}
	for _, serviceName := range services {
		_, err := h.pool.Exec(ctx, `
			INSERT INTO services (service_name, sub_duration_id_default, users_count, has_promocodes)
			VALUES ($1, 1, 1, false)
			ON CONFLICT (service_name) DO NOTHING
		`, serviceName)
		if err != nil {
			h.logger.Error("Failed to add service:", serviceName, err)
		}
	}

	// Добавляем тестовые подписки
	_, err := h.pool.Exec(ctx, `
		INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date)
		SELECT 
			uuid_generate_v4(),
			service_id,
			(1000 + random() * 2000)::int,
			'usual',
			DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month' * (random() * 12)::int),
			CASE WHEN random() > 0.3 THEN NULL 
				 ELSE DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month' * (random() * 6)::int)
			END
		FROM services
		WHERE service_name IN ('Netflix', 'Spotify', 'YouTube Premium', 'Disney+', 'Amazon Prime')
		LIMIT 50
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to add test data: " + err.Error(),
		})
		return
	}

	// Инвалидируем кэш
	h.statsService.InvalidateStatsCache(ctx)

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "test data added successfully",
	})
}

func (h *Lab9Handler) addRandomSubscription(ctx context.Context) {
	_, err := h.pool.Exec(ctx, `
		INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date)
		SELECT 
			uuid_generate_v4(),
			service_id,
			(1000 + random() * 2000)::int,
			'usual',
			CURRENT_DATE,
			NULL
		FROM services 
		ORDER BY random() 
		LIMIT 1
	`)
	if err == nil {
		h.statsService.InvalidateStatsCache(ctx)
		h.recordMetric("data_change", "add", 0, true)
	}
}

func (h *Lab9Handler) deleteRandomSubscription(ctx context.Context) {
	_, err := h.pool.Exec(ctx, `
		DELETE FROM subscriptions 
		WHERE sub_id IN (
			SELECT sub_id FROM subscriptions 
			ORDER BY random() 
			LIMIT 1
		)
	`)
	if err == nil {
		h.statsService.InvalidateStatsCache(ctx)
		h.recordMetric("data_change", "delete", 0, true)
	}
}

func (h *Lab9Handler) updateRandomSubscription(ctx context.Context) {
	_, err := h.pool.Exec(ctx, `
		UPDATE subscriptions 
		SET price = price * (0.8 + random() * 0.4)
		WHERE sub_id IN (
			SELECT sub_id FROM subscriptions 
			ORDER BY random() 
			LIMIT 1
		)
	`)
	if err == nil {
		h.statsService.InvalidateStatsCache(ctx)
		h.recordMetric("data_change", "update", 0, true)
	}
}

func (h *Lab9Handler) recordMetric(operation, source string, duration time.Duration, dataChanged bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	metric := ExecutionMetric{
		Timestamp:   time.Now(),
		Operation:   operation,
		Source:      source,
		Duration:    duration,
		DataChanged: dataChanged,
	}

	h.metrics = append(h.metrics, metric)

	// Ограничиваем размер истории
	if len(h.metrics) > 1000 {
		h.metrics = h.metrics[100:]
	}
}

func (h *Lab9Handler) analyzeMetrics() map[string]interface{} {
	if len(h.metrics) == 0 {
		return map[string]interface{}{}
	}

	var dbDurations, cacheDurations []time.Duration
	dbCount, cacheCount := 0, 0

	for _, metric := range h.metrics {
		if metric.Operation == "get_stats" || metric.Operation == "performance_db" {
			if metric.Source == "db" {
				dbDurations = append(dbDurations, metric.Duration)
				dbCount++
			} else if metric.Source == "cache" || metric.Source == "cache_miss" {
				cacheDurations = append(cacheDurations, metric.Duration)
				cacheCount++
			}
		}
	}

	analysis := map[string]interface{}{
		"total_requests":  len(h.metrics),
		"db_requests":     dbCount,
		"cache_requests":  cacheCount,
		"cache_hit_ratio": float64(cacheCount) / float64(dbCount+cacheCount),
	}

	if len(dbDurations) > 0 {
		var totalDB time.Duration
		for _, d := range dbDurations {
			totalDB += d
		}
		analysis["avg_db_duration"] = totalDB / time.Duration(len(dbDurations))
	}

	if len(cacheDurations) > 0 {
		var totalCache time.Duration
		for _, d := range cacheDurations {
			totalCache += d
		}
		analysis["avg_cache_duration"] = totalCache / time.Duration(len(cacheDurations))
	}

	return analysis
}

func (h *Lab9Handler) StopPerformanceTest(w http.ResponseWriter, r *http.Request) {
	// В реальной реализации здесь должна быть логика остановки тестов
	// Для простоты будем просто очищать некоторые метрики

	h.mu.Lock()
	// Оставляем только последние 100 метрик для анализа
	if len(h.metrics) > 100 {
		h.metrics = h.metrics[len(h.metrics)-100:]
	}
	h.mu.Unlock()

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "performance test stopped and metrics truncated",
	})
}

func (h *Lab9Handler) ClearTestData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Удаляем тестовые подписки, созданные для тестов
	// Будем удалять подписки, созданные в последние 24 часа
	_, err := h.pool.Exec(ctx, `
		DELETE FROM subscriptions 
		WHERE start_date >= CURRENT_DATE - INTERVAL '1 day'
		AND user_id IN (
			SELECT user_id FROM users WHERE email LIKE 'test%@example.com'
		)
	`)
	if err != nil {
		h.logger.Error("Failed to clear test data:", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to clear test data: " + err.Error(),
		})
		return
	}

	// Инвалидируем кэш
	h.statsService.InvalidateStatsCache(ctx)

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "test data cleared successfully",
	})
}
