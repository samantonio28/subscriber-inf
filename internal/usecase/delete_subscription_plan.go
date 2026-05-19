package usecase

import (
	"context"
	"strconv"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeleteSubscriptionPlanUC struct {
	planRepo  domain.SubscriptionPlanRepository
	planCache domain.SubscriptionPlanCache
	logger    logger.Logger
}

func NewDeleteSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, planCache domain.SubscriptionPlanCache, logger logger.Logger) (*DeleteSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if planCache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeleteSubscriptionPlanUC{planRepo: planRepo, planCache: planCache, logger: logger}, nil
}

func (uc *DeleteSubscriptionPlanUC) Delete(ctx context.Context, id domain.PlanID) error {
	uc.logger.Debug("deleting subscription plan", "plan_id", id, "source", "database")
	err := uc.planRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("failed to delete subscription plan from database", "plan_id", id, "error", err, "source", "database")
		return err
	}
	// Инвалидируем кэш
	cacheKey := strconv.Itoa(int(id))
	uc.logger.Debug("invalidating cache for deleted subscription plan", "plan_id", id, "cache_key", cacheKey)
	err = uc.planCache.DeleteSubscriptionPlan(ctx, cacheKey)
	if err != nil {
		uc.logger.Warn("failed to invalidate cache for subscription plan", "planID", id, "error", err)
	} else {
		uc.logger.Info("cache invalidated successfully", "plan_id", id, "cache_key", cacheKey)
	}
	uc.logger.Info("subscription plan deleted successfully", "plan_id", id, "source", "database")
	return nil
}