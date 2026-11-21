package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

// VerifyAuthCodeUC — второй фактор аутентификации: подтверждение входа
// одноразовым кодом, отправленным на email.
type VerifyAuthCodeUC struct {
	cache  domain.AuthCache
	logger logger.Logger
}

func NewVerifyAuthCodeUC(cache domain.AuthCache, logger logger.Logger) (*VerifyAuthCodeUC, error) {
	if cache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &VerifyAuthCodeUC{cache: cache, logger: logger}, nil
}

type VerifyCodeDTO struct {
	UserID uuid.UUID
	Code   string
}

// Verify сверяет код из запроса с кодом, сохранённым в Redis при входе.
// Верные коды одноразовые: после подтверждения код удаляется.
func (u *VerifyAuthCodeUC) Verify(ctx context.Context, dto VerifyCodeDTO) error {
	if dto.Code == "" {
		return fmt.Errorf("%w: code is required", domain.ErrInvalidInput)
	}

	locked, err := u.cache.Exists(ctx, domain.AuthLockKey(dto.UserID))
	if err != nil {
		return fmt.Errorf("check lock state: %w", err)
	}
	if locked {
		return domain.ErrAccountLocked
	}

	stored, err := u.cache.Get(ctx, domain.AuthCodeKey(dto.UserID))
	if err != nil {
		u.logger.Info("verify: no active code", "user_id", dto.UserID)
		return domain.ErrInvalidCode
	}
	if stored != dto.Code {
		u.logger.Info("verify: wrong code", "user_id", dto.UserID)
		return domain.ErrInvalidCode
	}
	// Код одноразовый.
	if err := u.cache.Delete(ctx, domain.AuthCodeKey(dto.UserID)); err != nil {
		u.logger.Warn("failed to delete used code", "error", err, "user_id", dto.UserID)
	}
	u.logger.Info("verify: second factor confirmed", "user_id", dto.UserID)
	return nil
}
