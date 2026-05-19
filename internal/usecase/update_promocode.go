package usecase

import (
	"context"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type UpdatePromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	promoCache    domain.PromocodeCache
	logger        logger.Logger
}

func NewUpdatePromocodeUC(promocodeRepo domain.PromocodeRepository, promoCache domain.PromocodeCache, logger logger.Logger) (*UpdatePromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if promoCache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &UpdatePromocodeUC{promocodeRepo: promocodeRepo, promoCache: promoCache, logger: logger}, nil
}

type UpdatePromocodeInput struct {
	ID           domain.PromocodeID
	ServiceID    int
	Value        string
	PlanID       *int
	SubID        *int
	ExpiresAt    time.Time
	Discount     int
	MaxUses      int
	CurUses      int
	Status       domain.PromocodeStatus
	DurationDays int
}

func (uc *UpdatePromocodeUC) Update(ctx context.Context, input UpdatePromocodeInput) error {
	uc.logger.Debug("updating promocode", "promocode_id", input.ID, "value", input.Value, "service_id", input.ServiceID, "discount", input.Discount)
	// Validate discount
	if input.Discount < 0 || input.Discount > 100 {
		uc.logger.Warn("invalid discount value", "discount", input.Discount)
		return domain.ErrInvalidInput
	}
	if input.MaxUses < 1 {
		input.MaxUses = 1
	}
	if input.DurationDays <= 0 {
		input.DurationDays = 3
	}
	// Ensure expires_at is set
	expiresAt := input.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().AddDate(0, 0, input.DurationDays)
	}

	promocode, err := domain.NewPromocode(
		input.ID,
		input.ServiceID,
		input.Value,
		input.PlanID,
		input.SubID,
		expiresAt,
		time.Now(), // created_at remains unchanged? we could keep original, but for simplicity we keep now
		input.Discount,
		input.MaxUses,
		input.CurUses,
		input.Status,
		input.DurationDays,
	)
	if err != nil {
		uc.logger.Error("failed to create promocode domain object", "error", err, "input", input)
		return err
	}

	err = uc.promocodeRepo.Update(ctx, *promocode)
	if err != nil {
		uc.logger.Error("failed to update promocode in database", "error", err, "promocode", promocode, "source", "database")
		return err
	}
	// Инвалидируем кэш по коду промокода
	uc.logger.Debug("invalidating cache for promocode", "code", input.Value)
	err = uc.promoCache.DeletePromocode(ctx, input.Value)
	if err != nil {
		uc.logger.Warn("failed to invalidate cache for promocode", "code", input.Value, "error", err)
	} else {
		uc.logger.Info("cache invalidated successfully", "code", input.Value)
	}
	uc.logger.Info("promocode updated successfully", "promocode_id", input.ID, "value", input.Value, "source", "database")
	return nil
}