package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type CreateSubUC struct {
	subR domain.SubscriptionRepository
}

func NewCreateSubUC(subR domain.SubscriptionRepository) (*CreateSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &CreateSubUC{subR: subR}, nil
}

func (u *CreateSubUC) NewSub(ctx context.Context, input SubscriptionDTO) (int, error) {
	sub, err := DTOToSub(input)
	if err != nil {
		return 0, err
	}
	sub.UserID = uuid.New()
	subId, err := u.subR.StoreSub(ctx, sub)
	if err != nil {
		return 0, err
	}
	return int(subId), nil
}
