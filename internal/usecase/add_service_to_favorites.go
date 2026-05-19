package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type AddServiceToFavoritesUC struct {
	userServiceRepo domain.UserServiceRepository
	logger          logger.Logger
}

func NewAddServiceToFavoritesUC(
	userServiceRepo domain.UserServiceRepository,
	logger logger.Logger,
) (*AddServiceToFavoritesUC, error) {
	if userServiceRepo == nil {
		return nil, domain.ErrInvalidUserServiceRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &AddServiceToFavoritesUC{userServiceRepo: userServiceRepo, logger: logger}, nil
}

func (u *AddServiceToFavoritesUC) Add(ctx context.Context, userID uuid.UUID, serviceID int) error {
	u.logger.Debug("adding service to favorites", "user_id", userID, "service_id", serviceID)

	// Проверяем, существует ли уже связь
	exists, err := u.userServiceRepo.Exists(ctx, userID, serviceID)
	if err != nil {
		u.logger.Error("failed to check existence", "error", err, "user_id", userID, "service_id", serviceID)
		return err
	}
	if exists {
		u.logger.Info("service already in favorites", "user_id", userID, "service_id", serviceID)
		return nil // уже добавлено, считаем успехом
	}

	err = u.userServiceRepo.Add(ctx, userID, serviceID)
	if err != nil {
		u.logger.Error("failed to add service to favorites", "error", err, "user_id", userID, "service_id", serviceID)
		return err
	}

	u.logger.Info("service added to favorites successfully", "user_id", userID, "service_id", serviceID)
	return nil
}