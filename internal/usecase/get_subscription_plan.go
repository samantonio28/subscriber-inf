package usecase

import (
	"context"
	"strconv"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetSubscriptionPlanUC struct {
	planRepo  domain.SubscriptionPlanRepository
	planCache domain.SubscriptionPlanCache
	logger    logger.Logger
}

func NewGetSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, planCache domain.SubscriptionPlanCache, logger logger.Logger) (*GetSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if planCache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetSubscriptionPlanUC{planRepo: planRepo, planCache: planCache, logger: logger}, nil
}

func (uc *GetSubscriptionPlanUC) ByID(ctx context.Context, id domain.PlanID) (domain.SubscriptionPlan, error) {
	// Пытаемся получить из кэша
	cacheKey := strconv.Itoa(int(id))
	uc.logger.Debug("attempting to get subscription plan from cache", "plan_id", id, "cache_key", cacheKey)
	plan, err := uc.planCache.GetSubscriptionPlan(ctx, cacheKey)
	if err == nil {
		uc.logger.Info("subscription plan cache hit", "plan_id", id, "cache_key", cacheKey, "source", "cache")
		return plan, nil
	}
	uc.logger.Debug("subscription plan cache miss", "plan_id", id, "cache_key", cacheKey, "error", err)
	// Если в кэше нет, идём в репозиторий
	uc.logger.Debug("fetching subscription plan from database", "plan_id", id, "source", "database")
	plan, err = uc.planRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get subscription plan from database", "plan_id", id, "error", err, "source", "database")
		return domain.SubscriptionPlan{}, err
	}
	// Сохраняем в кэш
	uc.logger.Debug("caching subscription plan", "plan_id", id, "cache_key", cacheKey, "ttl", "default")
	err = uc.planCache.SetSubscriptionPlan(ctx, cacheKey, plan, 0)
	if err != nil {
		uc.logger.Warn("subscription plan was not saved in cache, continuing", "plan_id", id, "error", err)
	} else {
		uc.logger.Info("subscription plan retrieved from db and cached", "plan_id", id, "cache_key", cacheKey, "source", "database")
	}
	return plan, nil
}

func (uc *GetSubscriptionPlanUC) ByServiceID(ctx context.Context, serviceID int) ([]domain.SubscriptionPlan, error) {
	uc.logger.Debug("fetching subscription plans by service", "service_id", serviceID, "source", "database")
	plans, err := uc.planRepo.GetByService(ctx, serviceID)
	if err != nil {
		uc.logger.Error("failed to get subscription plans by service", "service_id", serviceID, "error", err, "source", "database")
		return nil, err
	}
	uc.logger.Info("subscription plans retrieved by service", "service_id", serviceID, "count", len(plans), "source", "database")
	return plans, nil
}
