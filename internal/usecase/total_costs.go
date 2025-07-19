package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type TotalCostsUC struct {
	subR domain.SubscriptionRepository
}

func NewTotalCostsUC(subR domain.SubscriptionRepository) (*TotalCostsUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &TotalCostsUC{subR: subR}, nil
}

func (u *TotalCostsUC) TotalCosts(ctx context.Context, input SubsFilterDTO) (int, []int, error) {
	f, err := DTOToFilter(input)
	if err != nil {
		return 0, nil, err
	}

	sum, subIds, err := u.subR.SubsTotalCosts(ctx, f)
	if err != nil {
		return 0, nil, err
	}
	subIdsI := make([]int, 0, len(subIds))
	for _, s := range subIds {
		subIdsI = append(subIdsI, int(s))
	}
	return sum, subIdsI, nil
}
