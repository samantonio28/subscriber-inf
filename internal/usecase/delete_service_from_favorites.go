package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type DeleteServiceFromFavoritesUC struct {
	userServiceRepo domain.UserServiceRepository
	logger          logger.Logger
}

func NewDeleteServiceFromFavoritesUC(
	userServiceRepo domain.UserServiceRepository,
	logger logger.Logger,
) (*DeleteServiceFromFavoritesUC, error) {
	if userServiceRepo == nil {
		return nil, domain.ErrInvalidUserServiceRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &DeleteServiceFromFavoritesUC{userServiceRepo: userServiceRepo, logger: logger}, nil
}

func (u *DeleteServiceFromFavoritesUC) Delete(ctx context.Context, userID uuid.UUID, serviceID int) error {
	u.logger.Debug("deleting service from favorites", "user_id", userID, "service_id", serviceID)

	// Проверяем, существует ли связь
	exists, err := u.userServiceRepo.Exists(ctx, userID, serviceID)
	if err != nil {
		u.logger.Error("failed to check existence", "error", err, "user_id", userID, "service_id", serviceID)
		return err
	}
	if !exists {
		u.logger.Info("service not in favorites, nothing to delete", "user_id", userID, "service_id", serviceID)
		return nil // ничего удалять не нужно
	}

	err = u.userServiceRepo.Remove(ctx, userID, serviceID)
	if err != nil {
		u.logger.Error("failed to delete service from favorites", "error", err, "user_id", userID, "service_id", serviceID)
		return err
	}

	u.logger.Info("service deleted from favorites successfully", "user_id", userID, "service_id", serviceID)
	return nil
}