package usecase

import (
	"context"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type StatsOverviewUC struct {
	subRepo      domain.SubscriptionRepository
	promoRepo    domain.PromocodeRepository
	statsService domain.StatsService
	logger       logger.Logger
}

func NewStatsOverviewUC(subRepo domain.SubscriptionRepository, promoRepo domain.PromocodeRepository, statsService domain.StatsService, logger logger.Logger) (*StatsOverviewUC, error) {
	if subRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if promoRepo == nil {
		return nil, domain.ErrInvalidPromoRepo
	}
	if statsService == nil {
		return nil, domain.ErrInvalidStatsService
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &StatsOverviewUC{
		subRepo:      subRepo,
		promoRepo:    promoRepo,
		statsService: statsService,
		logger:       logger,
	}, nil
}

type StatsOverviewOutput struct {
	TotalSubscriptions   int            `json:"total_subscriptions"`
	ActiveSubscriptions  int            `json:"active_subscriptions"`
	TotalPromocodes      int            `json:"total_promocodes"`
	ActivePromocodes     int            `json:"active_promocodes"`
	TotalRevenue         int            `json:"total_revenue"`
	AvgSubscriptionPrice float64        `json:"avg_subscription_price"`
	MostPopularService   string         `json:"most_popular_service"`
	ServiceStats         []ServiceStats `json:"service_stats"`
}

type ServiceStats struct {
	ServiceName         string  `json:"service_name"`
	TotalSubscriptions  int     `json:"total_subscriptions"`
	TotalRevenue        int     `json:"total_revenue"`
	AvgSubscriptionCost float64 `json:"avg_subscription_cost"`
	ExecutionTime       string  `json:"execution_time"`
	Source              string  `json:"source"`
}

func (uc *StatsOverviewUC) GetOverview(ctx context.Context) (*StatsOverviewOutput, error) {
	uc.logger.Info("StatsOverviewUC.GetOverview called")

	// 1. Get service stats from StatsService
	stats, err := uc.statsService.GetServiceStatsFromDB(ctx)
	if err != nil {
		uc.logger.Error("GetOverview: failed to get service stats", "error", err)
		return nil, err
	}

	// 2. Get all promocodes
	promocodes, err := uc.promoRepo.GetAll(ctx)
	if err != nil {
		uc.logger.Error("GetOverview: failed to get promocodes", "error", err)
		return nil, err
	}

	// 3. Calculate promocode stats
	totalPromocodes := len(promocodes)
	activePromocodes := 0
	for _, pc := range promocodes {
		if pc.Status == domain.PromocodeStatusActive && pc.ExpiresAt.After(time.Now()) {
			activePromocodes++
		}
	}

	// 4. Calculate subscription stats from stats
	totalSubscriptions := stats.TotalSubs
	activeSubscriptions := 0
	totalRevenue := stats.TotalRevenue
	for _, ss := range stats.ServiceStats {
		activeSubscriptions += ss.ActiveSubs
	}

	// 5. Determine most popular service (by subscription count)
	mostPopularService := ""
	maxSubs := 0
	for _, ts := range stats.TopServices {
		if ts.SubCount > maxSubs {
			maxSubs = ts.SubCount
			mostPopularService = ts.ServiceName
		}
	}

	// 6. Calculate average subscription price
	avgSubscriptionPrice := 0.0
	if totalSubscriptions > 0 {
		avgSubscriptionPrice = float64(totalRevenue) / float64(totalSubscriptions)
	}

	// 7. Convert service stats to our format
	serviceStats := make([]ServiceStats, 0, len(stats.ServiceStats))
	for _, ss := range stats.ServiceStats {
		avgCost := 0.0
		if ss.TotalSubs > 0 {
			avgCost = float64(ss.TotalRevenue) / float64(ss.TotalSubs)
		}
		serviceStats = append(serviceStats, ServiceStats{
			ServiceName:         ss.ServiceName,
			TotalSubscriptions:  ss.TotalSubs,
			TotalRevenue:        ss.TotalRevenue,
			AvgSubscriptionCost: avgCost,
			ExecutionTime:       stats.ExecutionTime.String(),
			Source:              stats.Source,
		})
	}

	output := &StatsOverviewOutput{
		TotalSubscriptions:   totalSubscriptions,
		ActiveSubscriptions:  activeSubscriptions,
		TotalPromocodes:      totalPromocodes,
		ActivePromocodes:     activePromocodes,
		TotalRevenue:         totalRevenue,
		AvgSubscriptionPrice: avgSubscriptionPrice,
		MostPopularService:   mostPopularService,
		ServiceStats:         serviceStats,
	}
	return output, nil
}

type ReferralOverviewOutput struct {
	ReferralStatsItems []ReferralStatItem `json:"referral_stats_items"`
}

type ReferralStatItem struct {
	ReferrerID                 string  `json:"referrer_id"`
	ReferrerName               string  `json:"referrer_name"`
	ReferredCount              int     `json:"referred_count"`
	ConvertedToPurchase        int     `json:"converted_to_purchase"`
	AvgSubscriptionsPerReferred float64 `json:"avg_subscriptions_per_referred"`
}

func (uc *StatsOverviewUC) GetReferralOverview(ctx context.Context) (*ReferralOverviewOutput, error) {
	uc.logger.Info("StatsOverviewUC.GetReferralOverview called")

	referralStats, err := uc.statsService.GetReferralStatsFromDB(ctx)
	if err != nil {
		uc.logger.Error("GetReferralOverview: failed to get referral stats", "error", err)
		return nil, err
	}

	items := make([]ReferralStatItem, 0, len(referralStats))
	for _, rs := range referralStats {
		items = append(items, ReferralStatItem{
			ReferrerID:                 rs.ReferrerID.String(),
			ReferrerName:               rs.ReferrerName,
			ReferredCount:              rs.ReferredCount,
			ConvertedToPurchase:        rs.ConvertedToPurchase,
			AvgSubscriptionsPerReferred: rs.AvgSubscriptionsPerReferred,
		})
	}

	return &ReferralOverviewOutput{
		ReferralStatsItems: items,
	}, nil
}
