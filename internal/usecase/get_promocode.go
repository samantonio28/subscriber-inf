package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetPromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	logger        *logger.LogrusLogger
}

func NewGetPromocodeUC(promocodeRepo domain.PromocodeRepository, logger *logger.LogrusLogger) (*GetPromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetPromocodeUC{promocodeRepo: promocodeRepo, logger: logger}, nil
}

func (uc *GetPromocodeUC) ByID(ctx context.Context, id domain.PromocodeID) (domain.Promocode, error) {
	promocode, err := uc.promocodeRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return domain.Promocode{}, err
	}
	return promocode, nil
}

func (uc *GetPromocodeUC) ByCode(ctx context.Context, code string) (domain.Promocode, error) {
	promocode, err := uc.promocodeRepo.GetByCode(ctx, code)
	if err != nil {
		uc.logger.WithFields(map[string]any{"error": err})
		return domain.Promocode{}, err
	}
	return promocode, nil
}