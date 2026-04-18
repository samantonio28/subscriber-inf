package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeleteSubUC struct {
	subR   domain.SubscriptionRepository
	logger logger.Logger
}

func NewDeleteSubUC(subR domain.SubscriptionRepository, logger logger.Logger) (*DeleteSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeleteSubUC{subR: subR, logger: logger}, nil
}

func (u *DeleteSubUC) DeleteSub(ctx context.Context, subId int) error {
	u.logger.Debug("deleting subscription", "subId", subId)
	err := u.subR.DeleteSub(ctx, domain.SubID(subId))
	if err != nil {
		u.logger.Error("failed to delete subscription", "subId", subId, "error", err)
		return err
	}
	u.logger.Info("subscription deleted successfully", "subId", subId)
	return nil
}
