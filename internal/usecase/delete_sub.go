package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type DeleteSubUC struct {
	subR domain.SubscriptionRepository
}

func NewDeleteSubUC(subR domain.SubscriptionRepository) (*DeleteSubUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &DeleteSubUC{subR: subR}, nil
}

func (u *DeleteSubUC) DeleteSub(ctx context.Context, subId int) error {
	err := u.subR.DeleteSub(ctx, domain.SubID(subId))
	return err
}
