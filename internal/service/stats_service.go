package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
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
	TopServices   []TopService   `json:"top_services"`
	ServiceStats  []ServiceStats `json:"service_stats"`
	TotalRevenue  int            `json:"total_revenue"`
	TotalSubs     int            `json:"total_subs"`
	GeneratedAt   time.Time      `json:"generated_at"`
	ExecutionTime time.Duration  `json:"execution_time"`
	Source        string         `json:"source"` // "db" or "cache"
}

func NewStatsService(db *pgxpool.Pool, redisClient *redis.Client) (*StatsService, error) {
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
		LEFT JOIN subscription_plans sp ON s.service_id = sp.service_id
		LEFT JOIN subscriptions sub ON sp.plan_id = sub.plan_id
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
		LEFT JOIN subscription_plans sp ON s.service_id = sp.service_id
		LEFT JOIN subscriptions sub ON sp.plan_id = sub.plan_id
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

	// Преобразование во внутренние типы домена
	domainTopServices := make([]domain.TopService, len(topServices))
	for i, ts := range topServices {
		domainTopServices[i] = domain.TopService{
			ServiceName: ts.ServiceName,
			SubCount:    ts.SubCount,
			Revenue:     ts.Revenue,
		}
	}

	domainServiceStats := make([]domain.ServiceStats, len(serviceStats))
	for i, ss := range serviceStats {
		domainServiceStats[i] = domain.ServiceStats{
			ServiceName:  ss.ServiceName,
			TotalSubs:    ss.TotalSubs,
			ActiveSubs:   ss.ActiveSubs,
			TotalRevenue: ss.TotalRevenue,
		}
	}

	return &domain.StatsResponse{
		TopServices:   domainTopServices,
		ServiceStats:  domainServiceStats,
		TotalRevenue:  totalRevenue,
		TotalSubs:     totalSubs,
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Source:        "db",
	}, nil
}

// GetReferralStatsFromDB получает статистику по рефералам из VIEW referral_statistics
func (s *StatsService) GetReferralStatsFromDB(ctx context.Context) ([]domain.ReferralStat, error) {
	query := `
		SELECT
			referrer_id,
			referrer_name,
			referred_count,
			converted_to_purchase,
			COALESCE(avg_subscriptions_per_referred, 0) as avg_subscriptions_per_referred
		FROM referral_statistics
		ORDER BY referred_count DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query referral statistics: %w", err)
	}
	defer rows.Close()

	var stats []domain.ReferralStat
	for rows.Next() {
		var rs domain.ReferralStat
		if err := rows.Scan(&rs.ReferrerID, &rs.ReferrerName, &rs.ReferredCount, &rs.ConvertedToPurchase, &rs.AvgSubscriptionsPerReferred); err != nil {
			return nil, fmt.Errorf("failed to scan referral stat: %w", err)
		}
		stats = append(stats, rs)
	}
	return stats, nil
}

// GetServiceStatsFromCache получает статистику из кэша Redis
func (s *StatsService) GetServiceStatsFromCache(ctx context.Context) (*domain.StatsResponse, error) {
	start := time.Now()

	// Проверяем кэш
	cached, err := s.redis.Get(ctx, "service_stats")
	if err == nil && cached != "" {
		var stats domain.StatsResponse
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
