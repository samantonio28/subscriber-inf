package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type UpdateSubUC struct {
	subR domain.SubscriptionRepository
}

func NewUpdateSubUC(subR domain.SubscriptionRepository) (*UpdateSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &UpdateSubUC{subR: subR}, nil
}

func (u *UpdateSubUC) UpdateSub(ctx context.Context, subId int, input SubscriptionDTO) error {
	s, err := DTOToSub(input)
	if err != nil {
		return err
	}
	s.SubId = domain.SubID(subId)
	err = u.subR.UpdateSub(ctx, s)
	return err
}
