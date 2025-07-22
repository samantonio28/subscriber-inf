package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetSubUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewGetSubUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*GetSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetSubUC{subR: subR, logger: logger}, nil
}

func (u *GetSubUC) SubById(ctx context.Context, subId int) (SubscriptionDTO, error) {
	u.logger.Info("getting subscription by id", subId)
	sub, err := u.subR.Sub(ctx, domain.SubID(subId))
	if err != nil {
		u.logger.Error("error getting subscription by id", subId, err)
		return SubscriptionDTO{}, err
	}
	u.logger.Info("got subscription by id", subId, ": ", sub)
	return SubToDTO(sub), nil
}
