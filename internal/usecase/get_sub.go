package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetSubUC struct {
	subR   domain.SubscriptionRepository
	logger logger.Logger
}

func NewGetSubUC(subR domain.SubscriptionRepository, logger logger.Logger) (*GetSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetSubUC{subR: subR, logger: logger}, nil
}

func (u *GetSubUC) SubById(ctx context.Context, subId int) (SubscriptionDTO, error) {
	u.logger.Info("getting subscription by id", "subscription_id", subId, "source", "database")
	sub, err := u.subR.Sub(ctx, domain.SubID(subId))
	if err != nil {
		u.logger.Error("error getting subscription by id", "subscription_id", subId, "error", err, "source", "database")
		return SubscriptionDTO{}, err
	}
	u.logger.Info("subscription retrieved successfully", "subscription_id", subId, "user_id", sub.UserID, "service", sub.ServiceName, "source", "database")
	return SubToDTO(sub), nil
}
