package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeleteSubUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewDeleteSubUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*DeleteSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeleteSubUC{subR: subR, logger: logger}, nil
}

func (u *DeleteSubUC) DeleteSub(ctx context.Context, subId int) error {
	err := u.subR.DeleteSub(ctx, domain.SubID(subId))
	if err != nil {
		u.logger.WithFields(map[string]interface{}{
			"subId": subId,
			"error": err,
		}).Logger.Error("failed to delete subscription")
		return err
	}
	u.logger.Info("subscription", subId, "deleted")
	return nil
}
