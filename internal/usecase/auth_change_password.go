package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

// ChangePasswordUC — плановая смена пароля: пользователь, уже прошедший
// аутентификацию (X-User-ID), подтверждает знание текущего пароля и задаёт новый.
type ChangePasswordUC struct {
	userR  domain.UserRepository
	cache  domain.AuthCache
	logger logger.Logger
}

func NewChangePasswordUC(userR domain.UserRepository, cache domain.AuthCache, logger logger.Logger) (*ChangePasswordUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if cache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &ChangePasswordUC{userR: userR, cache: cache, logger: logger}, nil
}

type ChangePasswordDTO struct {
	UserID        uuid.UUID
	CurrentSecret string
	NewSecret     string
}

// Change проверяет текущий пароль и сохраняет bcrypt-хэш нового.
// Попутно инвалидирует активный код 2FA (если вход не был подтверждён).
func (u *ChangePasswordUC) Change(ctx context.Context, dto ChangePasswordDTO) error {
	if len(dto.NewSecret) < domain.MinSecretLen {
		return fmt.Errorf("%w: %w", domain.ErrInvalidInput, domain.ErrWeakSecret)
	}

	user, err := u.userR.GetUser(ctx, dto.UserID)
	if err != nil {
		u.logger.Error("change secret: user not found", "error", err, "user_id", dto.UserID)
		return err
	}

	// Плановая смена требует знания текущего пароля.
	if !CheckSecret(user.Password, dto.CurrentSecret) {
		u.logger.Info("change secret: wrong current secret", "user_id", dto.UserID)
		return domain.ErrInvalidCredentials
	}

	hash, err := HashSecret(dto.NewSecret)
	if err != nil {
		u.logger.Error("failed to hash new secret", "error", err)
		return err
	}
	user.Password = hash
	if err := u.userR.UpdateUser(ctx, user); err != nil {
		u.logger.Error("failed to update user secret", "error", err, "user_id", dto.UserID)
		return err
	}

	// Активный (неподтверждённый) код 2FA больше не действителен.
	_ = u.cache.Delete(ctx, domain.AuthCodeKey(user.UserID))
	u.logger.Info("secret changed", "user_id", user.UserID)
	return nil
}
