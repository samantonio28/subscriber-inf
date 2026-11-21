package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

// RegisterAuthUC — регистрация пользователя для входа по email + паролю.
// Пароль сохраняется только как bcrypt-хэш; роль — user.
type RegisterAuthUC struct {
	userR  domain.UserRepository
	logger logger.Logger
}

func NewRegisterAuthUC(userR domain.UserRepository, logger logger.Logger) (*RegisterAuthUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &RegisterAuthUC{userR: userR, logger: logger}, nil
}

type RegisterAuthDTO struct {
	Email        string
	UserName     string
	Age          int
	Balance      int
	ReferralCode *string
	Secret       string
}

func (u *RegisterAuthUC) Register(ctx context.Context, dto RegisterAuthDTO) (uuid.UUID, error) {
	u.logger.Debug("registering auth user", "email", dto.Email)

	if dto.Email == "" {
		return uuid.Nil, fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}
	if dto.UserName == "" {
		return uuid.Nil, fmt.Errorf("%w: user_name is required", domain.ErrInvalidInput)
	}
	if len(dto.Secret) < domain.MinSecretLen {
		return uuid.Nil, fmt.Errorf("%w: %w", domain.ErrInvalidInput, domain.ErrWeakSecret)
	}

	// Email должен быть свободен (штрафуем 409 Conflict).
	if _, err := u.userR.GetUserByEmail(ctx, dto.Email); err == nil {
		u.logger.Info("email already registered", "email", dto.Email)
		return uuid.Nil, domain.ErrEmailTaken
	}

	hash, err := HashSecret(dto.Secret)
	if err != nil {
		u.logger.Error("failed to hash secret", "error", err)
		return uuid.Nil, err
	}

	user := domain.User{
		UserID:       uuid.New(),
		Email:        dto.Email,
		Password:     hash,
		UserName:     dto.UserName,
		Age:          dto.Age,
		Balance:      dto.Balance,
		ReferralCode: dto.ReferralCode,
		Role:         domain.RoleUser,
	}
	if err := u.userR.StoreUser(ctx, user); err != nil {
		u.logger.Error("failed to store auth user", "error", err, "email", dto.Email)
		return uuid.Nil, err
	}
	u.logger.Info("auth user registered", "user_id", user.UserID, "email", dto.Email)
	return user.UserID, nil
}
