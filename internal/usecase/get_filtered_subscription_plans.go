package usecase

import (
	"context"
	"strings"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type SubscriptionPlanFilter struct {
	ServiceID        *int
	NameContains     *string
	PriceMin         *int
	PriceMax         *int
	DurationDaysMin  *int
	DurationDaysMax  *int
}

type GetFilteredSubscriptionPlansUC struct {
	planRepo domain.SubscriptionPlanRepository
	logger   logger.Logger
}

func NewGetFilteredSubscriptionPlansUC(planRepo domain.SubscriptionPlanRepository, logger logger.Logger) (*GetFilteredSubscriptionPlansUC, error) {
	if planRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetFilteredSubscriptionPlansUC{planRepo: planRepo, logger: logger}, nil
}

func (uc *GetFilteredSubscriptionPlansUC) GetFiltered(ctx context.Context, filter SubscriptionPlanFilter) ([]domain.SubscriptionPlan, error) {
	// Получить все планы
	all, err := uc.planRepo.GetAll(ctx)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return nil, err
	}

	// Применить фильтры
	var filtered []domain.SubscriptionPlan
	for _, plan := range all {
		if filter.ServiceID != nil && plan.ServiceID != *filter.ServiceID {
			continue
		}
		if filter.NameContains != nil && !strings.Contains(strings.ToLower(plan.Name), strings.ToLower(*filter.NameContains)) {
			continue
		}
		if filter.PriceMin != nil && plan.Price < *filter.PriceMin {
			continue
		}
		if filter.PriceMax != nil && plan.Price > *filter.PriceMax {
			continue
		}
		if filter.DurationDaysMin != nil && plan.DurationDays < *filter.DurationDaysMin {
			continue
		}
		if filter.DurationDaysMax != nil && plan.DurationDays > *filter.DurationDaysMax {
			continue
		}
		filtered = append(filtered, plan)
	}

	return filtered, nil
}