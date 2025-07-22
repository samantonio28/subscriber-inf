package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type UpdateSubUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewUpdateSubUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*UpdateSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &UpdateSubUC{subR: subR, logger: logger}, nil
}

func (u *UpdateSubUC) UpdateSub(ctx context.Context, subId int, input SubscriptionDTO) error {
	u.logger.Info("Updating subscription", subId)
	subToCheck, err := u.subR.Sub(ctx, domain.SubID(subId))
	if err != nil {
		u.logger.Error("not exists:", subId, err)
		return err
	}
	if input.StartDate.IsZero() {
		input.StartDate = subToCheck.StartDate
	}
	if input.ServiceName == " " {
		input.ServiceName = subToCheck.ServiceName
	}
	s, err := DTOToSub(input)
	if err != nil {
		u.logger.Error("invalid input:", input, err)
		return err
	}
	s.SubId = domain.SubID(subId)
	err = u.subR.UpdateSub(ctx, s)
	if err != nil {
		u.logger.Error("error updating subscription:", subId, err)
		return err
	}
	u.logger.Info("subscription updated:", subId)
	return nil
}
