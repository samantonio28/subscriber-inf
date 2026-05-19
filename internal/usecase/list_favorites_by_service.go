package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type ListFavoritesByServiceUC struct {
	userServiceRepo domain.UserServiceRepository
	logger          logger.Logger
}

func NewListFavoritesByServiceUC(
	userServiceRepo domain.UserServiceRepository,
	logger logger.Logger,
) (*ListFavoritesByServiceUC, error) {
	if userServiceRepo == nil {
		return nil, domain.ErrInvalidUserServiceRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &ListFavoritesByServiceUC{userServiceRepo: userServiceRepo, logger: logger}, nil
}

func (u *ListFavoritesByServiceUC) List(ctx context.Context, serviceID int) ([]uuid.UUID, error) {
	u.logger.Debug("listing favorites by service", "service_id", serviceID)

	userIDs, err := u.userServiceRepo.ListByService(ctx, serviceID)
	if err != nil {
		u.logger.Error("failed to list favorites by service", "error", err, "service_id", serviceID)
		return nil, err
	}

	u.logger.Info("favorites by service listed successfully", "service_id", serviceID, "count", len(userIDs))
	return userIDs, nil
}
