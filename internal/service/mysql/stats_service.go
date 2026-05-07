package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/redis"
)

type StatsService struct {
	db    *sql.DB
	redis *redis.Client
}

type ServiceStats struct {
	ServiceName  string `json:"service_name"`
	TotalSubs    int    `json:"total_subs"`
	ActiveSubs   int    `json:"active_subs"`
	TotalRevenue int    `json:"total_revenue"`
}

type TopService struct {
	ServiceName string `json:"service_name"`
	SubCount    int    `json:"sub_count"`
	Revenue     int    `json:"revenue"`
}

type StatsResponse struct {
	TopServices   []TopService   `json:"top_services"`
	ServiceStats  []ServiceStats `json:"service_stats"`
	TotalRevenue  int            `json:"total_revenue"`
	TotalSubs     int            `json:"total_subs"`
	GeneratedAt   time.Time      `json:"generated_at"`
	ExecutionTime time.Duration  `json:"execution_time"`
	Source        string         `json:"source"` // "db" or "cache"
}

func NewStatsService(db *sql.DB, redisClient *redis.Client) (*StatsService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &StatsService{
		db:    db,
		redis: redisClient,
	}, nil
}

// GetServiceStatsFromDB получает статистику напрямую из БД
func (s *StatsService) GetServiceStatsFromDB(ctx context.Context) (*domain.StatsResponse, error) {
	start := time.Now()

	// Топ 10 сервисов по количеству подписок
	topServicesQuery := `
		SELECT
			s.service_name,
			COUNT(sub.sub_id) as sub_count,
			COALESCE(SUM(sub.price), 0) as revenue
		FROM services s
		LEFT JOIN subscriptions sub ON s.service_id = sub.service_id
		WHERE sub.end_date IS NULL OR sub.end_date >= CURDATE()
		GROUP BY s.service_id, s.service_name
		ORDER BY sub_count DESC, revenue DESC
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, topServicesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query top services: %w", err)
	}
	defer rows.Close()

	var topServices []TopService
	for rows.Next() {
		var ts TopService
		if err := rows.Scan(&ts.ServiceName, &ts.SubCount, &ts.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan top service: %w", err)
		}
		topServices = append(topServices, ts)
	}

	// Общая статистика по сервисам
	statsQuery := `
		SELECT
			s.service_name,
			COUNT(sub.sub_id) as total_subs,
			COUNT(CASE WHEN sub.end_date IS NULL OR sub.end_date >= CURDATE() THEN 1 END) as active_subs,
			COALESCE(SUM(sub.price), 0) as total_revenue
		FROM services s
		LEFT JOIN subscriptions sub ON s.service_id = sub.service_id
		GROUP BY s.service_id, s.service_name
		ORDER BY total_revenue DESC
	`

	rows2, err := s.db.QueryContext(ctx, statsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query service stats: %w", err)
	}
	defer rows2.Close()

	var serviceStats []ServiceStats
	for rows2.Next() {
		var ss ServiceStats
		if err := rows2.Scan(&ss.ServiceName, &ss.TotalSubs, &ss.ActiveSubs, &ss.TotalRevenue); err != nil {
			return nil, fmt.Errorf("failed to scan service stats: %w", err)
		}
		serviceStats = append(serviceStats, ss)
	}

	// Общие итоги
	totalRevenue := 0
	totalSubs := 0
	for _, ss := range serviceStats {
		totalRevenue += ss.TotalRevenue
		totalSubs += ss.TotalSubs
	}

	executionTime := time.Since(start)

	// Попробуем получить из кэша
	cacheKey := "stats:overview"
	var cachedResponse *domain.StatsResponse
	if cachedData, err := s.redis.Get(ctx, cacheKey); err == nil && cachedData != "" {
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
			cachedResponse.Source = "cache"
			return cachedResponse, nil
		}
	}

	response := &domain.StatsResponse{
		TopServices:   convertTopServices(topServices),
		ServiceStats:  convertServiceStats(serviceStats),
		TotalRevenue:  totalRevenue,
		TotalSubs:     totalSubs,
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Source:        "db",
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(response); err == nil {
		_ = s.redis.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return response, nil
}

func convertTopServices(ts []TopService) []domain.TopService {
	result := make([]domain.TopService, len(ts))
	for i, t := range ts {
		result[i] = domain.TopService{
			ServiceName: t.ServiceName,
			SubCount:    t.SubCount,
			Revenue:     t.Revenue,
		}
	}
	return result
}

func convertServiceStats(ss []ServiceStats) []domain.ServiceStats {
	result := make([]domain.ServiceStats, len(ss))
	for i, s := range ss {
		result[i] = domain.ServiceStats{
			ServiceName:  s.ServiceName,
			TotalSubs:    s.TotalSubs,
			ActiveSubs:   s.ActiveSubs,
			TotalRevenue: s.TotalRevenue,
		}
	}
	return result
}