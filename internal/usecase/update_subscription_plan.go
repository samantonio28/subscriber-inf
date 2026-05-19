package usecase

import (
	"context"
	"strconv"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type UpdateSubscriptionPlanUC struct {
	planRepo  domain.SubscriptionPlanRepository
	planCache domain.SubscriptionPlanCache
	logger    logger.Logger
}

func NewUpdateSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, planCache domain.SubscriptionPlanCache, logger logger.Logger) (*UpdateSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if planCache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &UpdateSubscriptionPlanUC{planRepo: planRepo, planCache: planCache, logger: logger}, nil
}

type UpdateSubscriptionPlanInput struct {
	ID           domain.PlanID
	ServiceID    int
	Name         string
	DurationDays int
	Price        int
}

func (uc *UpdateSubscriptionPlanUC) Update(ctx context.Context, input UpdateSubscriptionPlanInput) error {
	uc.logger.Debug("updating subscription plan", "plan_id", input.ID, "service_id", input.ServiceID, "name", input.Name, "duration_days", input.DurationDays, "price", input.Price)
	plan, err := domain.NewSubscriptionPlan(
		input.ID,
		input.ServiceID,
		input.Name,
		input.DurationDays,
		input.Price,
	)
	if err != nil {
		uc.logger.Error("failed to create subscription plan domain object", "error", err, "input", input)
		return err
	}

	err = uc.planRepo.Update(ctx, *plan)
	if err != nil {
		uc.logger.Error("failed to update subscription plan in database", "error", err, "plan", plan, "source", "database")
		return err
	}
	// Инвалидируем кэш
	cacheKey := strconv.Itoa(int(input.ID))
	uc.logger.Debug("invalidating cache for subscription plan", "plan_id", input.ID, "cache_key", cacheKey)
	err = uc.planCache.DeleteSubscriptionPlan(ctx, cacheKey)
	if err != nil {
		uc.logger.Warn("failed to invalidate cache for subscription plan", "planID", input.ID, "error", err)
	} else {
		uc.logger.Info("cache invalidated successfully", "plan_id", input.ID, "cache_key", cacheKey)
	}
	uc.logger.Info("subscription plan updated successfully", "plan_id", input.ID, "source", "database")
	return nil
}