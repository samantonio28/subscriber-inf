package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type UpdateSubscriptionPlanUC struct {
	planRepo domain.SubscriptionPlanRepository
	logger   logger.Logger
}

func NewUpdateSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, logger logger.Logger) (*UpdateSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &UpdateSubscriptionPlanUC{planRepo: planRepo, logger: logger}, nil
}

type UpdateSubscriptionPlanInput struct {
	ID           domain.PlanID
	ServiceID    int
	Name         string
	DurationDays int
	Price        int
}

func (uc *UpdateSubscriptionPlanUC) Update(ctx context.Context, input UpdateSubscriptionPlanInput) error {
	plan, err := domain.NewSubscriptionPlan(
		input.ID,
		input.ServiceID,
		input.Name,
		input.DurationDays,
		input.Price,
	)
	if err != nil {
		return err
	}

	err = uc.planRepo.Update(ctx, *plan)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return err
	}
	uc.logger.Info("subscription plan updated")
	return nil
}