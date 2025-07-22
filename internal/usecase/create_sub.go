package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type CreateSubUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewCreateSubUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*CreateSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &CreateSubUC{subR: subR, logger: logger}, nil
}

func (u *CreateSubUC) NewSub(ctx context.Context, input SubscriptionDTO) (int, error) {
	sub, err := DTOToSub(input)
	if err != nil {
		u.logger.WithFields(map[string]any{"error": err})
		return 0, err
	}
	if sub.UserID == uuid.Nil {
		u.logger.Info("there was no user id")
		sub.UserID = uuid.New()
	}
	subId, err := u.subR.StoreSub(ctx, sub)
	if err != nil {
		u.logger.WithFields(map[string]any{"error": err})
		return 0, err
	}
	u.logger.Info("subscription created")
	return int(subId), nil
}
