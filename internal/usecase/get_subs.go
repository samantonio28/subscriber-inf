package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetSubsUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewGetSubsUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*GetSubsUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetSubsUC{subR: subR, logger: logger}, nil
}

func (u *GetSubsUC) SubsByUserId(ctx context.Context, userId uuid.UUID) ([]SubscriptionDTO, error) {
	u.logger.Info("getting subscriptions by user id", userId)
	subs, err := u.subR.UserSubs(ctx, userId)
	if err != nil {
		u.logger.Error("error getting subscriptions by user id", userId, err)
		return nil, err
	}
	dto := make([]SubscriptionDTO, 0, len(subs))
	for _, s := range subs {
		dto = append(dto, SubToDTO(s))
	}
	u.logger.Info("got subscriptions by user id", userId, ": ", len(dto))
	return dto, nil
}
