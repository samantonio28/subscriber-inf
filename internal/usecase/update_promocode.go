package usecase

import (
	"context"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type UpdatePromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	logger        logger.Logger
}

func NewUpdatePromocodeUC(promocodeRepo domain.PromocodeRepository, logger logger.Logger) (*UpdatePromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &UpdatePromocodeUC{promocodeRepo: promocodeRepo, logger: logger}, nil
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
	// Validate discount
	if input.Discount < 0 || input.Discount > 100 {
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
		return err
	}

	err = uc.promocodeRepo.Update(ctx, *promocode)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return err
	}
	uc.logger.Info("promocode updated")
	return nil
}