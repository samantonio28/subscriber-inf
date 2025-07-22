package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type TotalCostsUC struct {
	subR   domain.SubscriptionRepository
	logger *logger.LogrusLogger
}

func NewTotalCostsUC(subR domain.SubscriptionRepository, logger *logger.LogrusLogger) (*TotalCostsUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &TotalCostsUC{subR: subR, logger: logger}, nil
}

func (u *TotalCostsUC) TotalCosts(ctx context.Context, input SubsFilterDTO) (int, []int, error) {
	u.logger.Info("TotalCosts", "input", input)
	f, err := DTOToFilter(input)
	if err != nil {
		u.logger.Error("TotalCosts", "input", input, "error", err)
		return 0, nil, err
	}

	sum, subIds, err := u.subR.SubsTotalCosts(ctx, f)
	if err != nil {
		u.logger.Error("TotalCosts", "input", input, "error", err)
		return 0, nil, err
	}
	subIdsI := make([]int, 0, len(subIds))
	for _, s := range subIds {
		subIdsI = append(subIdsI, int(s))
	}
	u.logger.Info("TotalCosts", "input", input, "output", sum, "subIds len", len(subIdsI))
	return sum, subIdsI, nil
}
