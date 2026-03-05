package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetSubscriptionPlanUC struct {
	planRepo domain.SubscriptionPlanRepository
	logger   *logger.LogrusLogger
}

func NewGetSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, logger *logger.LogrusLogger) (*GetSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetSubscriptionPlanUC{planRepo: planRepo, logger: logger}, nil
}

func (uc *GetSubscriptionPlanUC) ByID(ctx context.Context, id domain.PlanID) (domain.SubscriptionPlan, error) {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return domain.SubscriptionPlan{}, err
	}
	uc.logger.Info("subscription plan retrieved")
	return plan, nil
}

func (uc *GetSubscriptionPlanUC) ByServiceID(ctx context.Context, serviceID int) ([]domain.SubscriptionPlan, error) {
	plans, err := uc.planRepo.GetByService(ctx, serviceID)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return nil, err
	}
	uc.logger.Info("subscription plans retrieved by service")
	return plans, nil
}