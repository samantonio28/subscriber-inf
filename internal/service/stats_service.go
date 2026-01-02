package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/redis"
)

type StatsService struct {
	db    *pgxpool.Pool
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
	TopServices    []TopService   `json:"top_services"`
	ServiceStats   []ServiceStats `json:"service_stats"`
	TotalRevenue   int            `json:"total_revenue"`
	TotalSubs      int            `json:"total_subs"`
	GeneratedAt    time.Time      `json:"generated_at"`
	ExecutionTime  time.Duration  `json:"execution_time"`
	Source         string         `json:"source"` // "db" or "cache"
}

func NewStatsService(db *pgxpool.Pool, redisClient *redis.Client) *StatsService {
	return &StatsService{
		db:    db,
		redis: redisClient,
	}
}

// GetServiceStatsFromDB получает статистику напрямую из БД
func (s *StatsService) GetServiceStatsFromDB(ctx context.Context) (*StatsResponse, error) {
	start := time.Now()

	// Топ 10 сервисов по количеству подписок
	topServicesQuery := `
		SELECT 
			s.service_name,
			COUNT(sub.sub_id) as sub_count,
			COALESCE(SUM(sub.price), 0) as revenue
		FROM services s
		LEFT JOIN subscriptions sub ON s.service_id = sub.service_id
		WHERE sub.end_date IS NULL OR sub.end_date >= CURRENT_DATE
		GROUP BY s.service_id, s.service_name
		ORDER BY sub_count DESC, revenue DESC
		LIMIT 10
	`

	rows, err := s.db.Query(ctx, topServicesQuery)
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
			COUNT(CASE WHEN sub.end_date IS NULL OR sub.end_date >= CURRENT_DATE THEN 1 END) as active_subs,
			COALESCE(SUM(sub.price), 0) as total_revenue
		FROM services s
		LEFT JOIN subscriptions sub ON s.service_id = sub.service_id
		GROUP BY s.service_id, s.service_name
		ORDER BY total_revenue DESC
	`

	rows, err = s.db.Query(ctx, statsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query service stats: %w", err)
	}
	defer rows.Close()

	var serviceStats []ServiceStats
	totalRevenue := 0
	totalSubs := 0

	for rows.Next() {
		var ss ServiceStats
		if err := rows.Scan(&ss.ServiceName, &ss.TotalSubs, &ss.ActiveSubs, &ss.TotalRevenue); err != nil {
			return nil, fmt.Errorf("failed to scan service stats: %w", err)
		}
		serviceStats = append(serviceStats, ss)
		totalRevenue += ss.TotalRevenue
		totalSubs += ss.TotalSubs
	}

	executionTime := time.Since(start)

	return &StatsResponse{
		TopServices:   topServices,
		ServiceStats:  serviceStats,
		TotalRevenue:  totalRevenue,
		TotalSubs:     totalSubs,
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Source:        "db",
	}, nil
}

// GetServiceStatsFromCache получает статистику из кэша Redis
func (s *StatsService) GetServiceStatsFromCache(ctx context.Context) (*StatsResponse, error) {
	start := time.Now()

	// Проверяем кэш
	cached, err := s.redis.Get(ctx, "service_stats")
	if err == nil && cached != "" {
		var stats StatsResponse
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			stats.ExecutionTime = time.Since(start)
			stats.Source = "cache"
			return &stats, nil
		}
	}

	// Если в кэше нет, получаем из БД и кэшируем
	stats, err := s.GetServiceStatsFromDB(ctx)
	if err != nil {
		return nil, err
	}

	// Кэшируем на 30 секунд
	statsJSON, err := json.Marshal(stats)
	if err == nil {
		s.redis.Set(ctx, "service_stats", statsJSON, 30*time.Second)
	}

	stats.ExecutionTime = time.Since(start)
	stats.Source = "cache_miss"
	return stats, nil
}

// InvalidateStatsCache очищает кэш статистики
func (s *StatsService) InvalidateStatsCache(ctx context.Context) error {
	return s.redis.Delete(ctx, "service_stats")
}
