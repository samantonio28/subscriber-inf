package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeleteSubscriptionPlanUC struct {
	planRepo domain.SubscriptionPlanRepository
	logger   logger.Logger
}

func NewDeleteSubscriptionPlanUC(planRepo domain.SubscriptionPlanRepository, logger logger.Logger) (*DeleteSubscriptionPlanUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeleteSubscriptionPlanUC{planRepo: planRepo, logger: logger}, nil
}

func (uc *DeleteSubscriptionPlanUC) Delete(ctx context.Context, id domain.PlanID) error {
	err := uc.planRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return err
	}
	uc.logger.Info("subscription plan deleted")
	return nil
}