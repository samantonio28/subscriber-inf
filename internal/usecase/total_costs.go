package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type TotalCostsUC struct {
	subR   domain.SubscriptionRepository
	logger logger.Logger
}

func NewTotalCostsUC(subR domain.SubscriptionRepository, logger logger.Logger) (*TotalCostsUC, error) {
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &TotalCostsUC{subR: subR, logger: logger}, nil
}

func (u *TotalCostsUC) TotalCosts(ctx context.Context, input SubsFilterDTO) (int, []domain.Subscription, error) {
	u.logger.Info("TotalCosts", "input", input)
	f, err := DTOToFilter(input)
	if err != nil {
		u.logger.Error("TotalCosts", "input", input, "error", err)
		return 0, nil, err
	}

	sum, subs, err := u.subR.SubsTotalCosts(ctx, f)
	if err != nil {
		u.logger.Error("TotalCosts", "input", input, "error", err)
		return 0, nil, err
	}
	u.logger.Info("TotalCosts", "input", input, "output", sum, "subs len", len(subs))
	return sum, subs, nil
}
