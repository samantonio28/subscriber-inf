package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type PromocodeFilter struct {
	ServiceID     *int
	PlanID        *int
	SubID         *int
	Status        *domain.PromocodeStatus
	DiscountMin   *int
	DiscountMax   *int
	ExpiresBefore *time.Time
	ExpiresAfter  *time.Time
	CodeContains  *string
}

type GetFilteredPromocodesUC struct {
	promocodeRepo domain.PromocodeRepository
	logger        logger.Logger
}

func NewGetFilteredPromocodesUC(promocodeRepo domain.PromocodeRepository, logger logger.Logger) (*GetFilteredPromocodesUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetFilteredPromocodesUC{promocodeRepo: promocodeRepo, logger: logger}, nil
}

func (uc *GetFilteredPromocodesUC) GetFiltered(ctx context.Context, filter PromocodeFilter) ([]domain.Promocode, error) {
	// Получить все промокоды из репозитория
	all, err := uc.promocodeRepo.GetAll(ctx)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return nil, err
	}

	// Применить фильтры
	var filtered []domain.Promocode
	for _, pc := range all {
		if filter.ServiceID != nil && pc.ServiceID != *filter.ServiceID {
			continue
		}
		if filter.PlanID != nil && (pc.PlanID == nil || *pc.PlanID != *filter.PlanID) {
			continue
		}
		if filter.SubID != nil && (pc.SubID == nil || *pc.SubID != *filter.SubID) {
			continue
		}
		if filter.Status != nil && pc.Status != *filter.Status {
			continue
		}
		if filter.DiscountMin != nil && pc.Discount < *filter.DiscountMin {
			continue
		}
		if filter.DiscountMax != nil && pc.Discount > *filter.DiscountMax {
			continue
		}
		if filter.ExpiresBefore != nil && (pc.ExpiresAt.IsZero() || !pc.ExpiresAt.Before(*filter.ExpiresBefore)) {
			continue
		}
		if filter.ExpiresAfter != nil && (pc.ExpiresAt.IsZero() || !pc.ExpiresAt.After(*filter.ExpiresAfter)) {
			continue
		}
		if filter.CodeContains != nil && !strings.Contains(strings.ToLower(pc.Value), strings.ToLower(*filter.CodeContains)) {
			continue
		}
		filtered = append(filtered, pc)
	}

	return filtered, nil
}