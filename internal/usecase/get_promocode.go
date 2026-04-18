package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetPromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	logger        logger.Logger
}

func NewGetPromocodeUC(promocodeRepo domain.PromocodeRepository, logger logger.Logger) (*GetPromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetPromocodeUC{promocodeRepo: promocodeRepo, logger: logger}, nil
}

func (uc *GetPromocodeUC) ByID(ctx context.Context, id domain.PromocodeID) (domain.Promocode, error) {
	uc.logger.Debug("fetching promocode by ID", "promocode_id", id)
	promocode, err := uc.promocodeRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to fetch promocode by ID", "promocode_id", id, "error", err)
		return domain.Promocode{}, err
	}
	uc.logger.Debug("promocode fetched successfully", "promocode_id", id)
	return promocode, nil
}

func (uc *GetPromocodeUC) ByCode(ctx context.Context, code string) (domain.Promocode, error) {
	uc.logger.Debug("fetching promocode by code", "code", code)
	promocode, err := uc.promocodeRepo.GetByCode(ctx, code)
	if err != nil {
		uc.logger.Error("failed to fetch promocode by code", "code", code, "error", err)
		return domain.Promocode{}, err
	}
	uc.logger.Debug("promocode fetched successfully", "code", code, "promocode_id", promocode.PromocodeID)
	return promocode, nil
}