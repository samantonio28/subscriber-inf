package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetUserUC struct {
	userR  domain.UserRepository
	logger logger.Logger
}

func NewGetUserUC(userR domain.UserRepository, logger logger.Logger) (*GetUserUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetUserUC{userR: userR, logger: logger}, nil
}

func (u *GetUserUC) UserById(ctx context.Context, userID uuid.UUID) (GetUserDTO, error) {
	u.logger.Info("getting user by id", userID)
	user, err := u.userR.GetUser(ctx, userID)
	if err != nil {
		u.logger.Error("error getting user by id", userID, err)
		return GetUserDTO{}, err
	}
	u.logger.Info("got user by id", userID, ": ", user)
	return UserToGetUserDTO(user), nil
}
