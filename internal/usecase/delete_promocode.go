package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeletePromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	logger        logger.Logger
}

func NewDeletePromocodeUC(promocodeRepo domain.PromocodeRepository, logger logger.Logger) (*DeletePromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeletePromocodeUC{promocodeRepo: promocodeRepo, logger: logger}, nil
}

func (uc *DeletePromocodeUC) Delete(ctx context.Context, id domain.PromocodeID) error {
	uc.logger.Debug("deleting promocode", "promocode_id", id)
	err := uc.promocodeRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("failed to delete promocode", "promocode_id", id, "error", err)
		return err
	}
	uc.logger.Info("promocode deleted successfully", "promocode_id", id)
	return nil
}