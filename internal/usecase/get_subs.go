package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type GetSubsUC struct {
	subR domain.SubscriptionRepository
}

func NewGetSubsUC(subR domain.SubscriptionRepository) (*GetSubsUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &GetSubsUC{subR: subR}, nil
}

func (u *GetSubsUC) SubsByUserId(ctx context.Context, userId uuid.UUID) ([]SubscriptionDTO, error) {
	subs, err := u.subR.UserSubs(ctx, userId)
	if err != nil {
		return nil, err
	}
	dto := make([]SubscriptionDTO, 0, len(subs))
	for _, s := range subs {
		dto = append(dto, SubToDTO(s))
	}
	return dto, nil
}
