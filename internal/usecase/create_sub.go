package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type CreateSubUC struct {
	subR   domain.SubscriptionRepository
	logger logger.Logger
}

func NewCreateSubUC(subR domain.SubscriptionRepository, logger logger.Logger) (*CreateSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &CreateSubUC{subR: subR, logger: logger}, nil
}

func (u *CreateSubUC) NewSub(ctx context.Context, input SubscriptionDTO) (int, error) {
	u.logger.Debug("creating new subscription", "input", input)
	sub, err := DTOToSub(input)
	if err != nil {
		u.logger.Error("failed to convert DTO to subscription", "error", err, "input", input)
		return 0, err
	}
	if sub.UserID == uuid.Nil {
		u.logger.Info("generating new user ID for subscription")
		sub.UserID = uuid.New()
	}
	subId, err := u.subR.StoreSub(ctx, sub)
	if err != nil {
		u.logger.Error("failed to store subscription", "error", err, "subscription", sub)
		return 0, err
	}
	u.logger.Info("subscription created successfully", "subscription_id", subId, "user_id", sub.UserID)
	return int(subId), nil
}
