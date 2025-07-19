package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type GetSubUC struct {
	subR domain.SubscriptionRepository
}

func NewGetSubUC(subR domain.SubscriptionRepository) (*GetSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &GetSubUC{subR: subR}, nil
}

func (u *GetSubUC) SubById(ctx context.Context, subId int) (SubscriptionDTO, error) {
	sub, err := u.subR.Sub(ctx, domain.SubID(subId))
	if err != nil {
		return SubscriptionDTO{}, err
	}
	return SubToDTO(*sub), nil
}
