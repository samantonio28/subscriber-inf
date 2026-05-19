package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type ListUserFavoritesUC struct {
	userServiceRepo domain.UserServiceRepository
	logger          logger.Logger
}

func NewListUserFavoritesUC(
	userServiceRepo domain.UserServiceRepository,
	logger logger.Logger,
) (*ListUserFavoritesUC, error) {
	if userServiceRepo == nil {
		return nil, domain.ErrInvalidUserServiceRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &ListUserFavoritesUC{userServiceRepo: userServiceRepo, logger: logger}, nil
}

func (u *ListUserFavoritesUC) List(ctx context.Context, userID uuid.UUID) ([]int, error) {
	u.logger.Debug("listing user favorites", "user_id", userID)

	serviceIDs, err := u.userServiceRepo.ListByUser(ctx, userID)
	if err != nil {
		u.logger.Error("failed to list user favorites", "error", err, "user_id", userID)
		return nil, err
	}

	u.logger.Info("user favorites listed successfully", "user_id", userID, "count", len(serviceIDs))
	return serviceIDs, nil
}