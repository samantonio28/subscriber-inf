package usecase

import (
	"context"
	"fmt"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

// RecoverAuthUC — сценарий восстановления доступа при блокировке:
//   1. Recover — заблокированный пользователь запрашивает код восстановления,
//      который отправляется на email (второй фактор).
//   2. Confirm — подтверждение кодом снимает блокировку, сбрасывает счётчик
//      неудачных попыток и задаёт новый пароль.
type RecoverAuthUC struct {
	userR  domain.UserRepository
	cache  domain.AuthCache
	mail   domain.Mailer
	logger logger.Logger
}

func NewRecoverAuthUC(userR domain.UserRepository, cache domain.AuthCache, mail domain.Mailer, logger logger.Logger) (*RecoverAuthUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if cache == nil {
		return nil, domain.ErrInvalidCache
	}
	if mail == nil {
		return nil, fmt.Errorf("mailer is not defined")
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &RecoverAuthUC{userR: userR, cache: cache, mail: mail, logger: logger}, nil
}

// Recover отправляет код восстановления на email пользователя.
func (u *RecoverAuthUC) Recover(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}
	user, err := u.userR.GetUserByEmail(ctx, email)
	if err != nil {
		u.logger.Info("recover: unknown email", "email", email)
		return domain.ErrUserNotFound
	}

	code, err := domain.GenerateAuthCode()
	if err != nil {
		u.logger.Error("failed to generate recovery code", "error", err)
		return fmt.Errorf("generate recovery code: %w", err)
	}
	if err := u.cache.Set(ctx, domain.AuthRecoveryCodeKey(email), code, domain.AuthCodeTTL); err != nil {
		u.logger.Error("failed to store recovery code", "error", err, "email", email)
		return fmt.Errorf("store recovery code: %w", err)
	}
	if err := u.mail.Send(ctx, user.Email, "Восстановление доступа", domain.SendRecoveryCodeLetter(code)); err != nil {
		u.logger.Error("failed to send recovery code", "error", err, "email", email)
		return fmt.Errorf("send recovery code: %w", err)
	}
	u.logger.Info("recover: code sent", "email", email)
	return nil
}

// Confirm проверяет код восстановления, снимает блокировку, сбрасывает счётчик
// неудачных попыток и устанавливает новый пароль.
func (u *RecoverAuthUC) Confirm(ctx context.Context, email, code, newSecret string) error {
	if email == "" || code == "" {
		return fmt.Errorf("%w: email and code are required", domain.ErrInvalidInput)
	}
	if len(newSecret) < domain.MinSecretLen {
		return fmt.Errorf("%w: %w", domain.ErrInvalidInput, domain.ErrWeakSecret)
	}

	user, err := u.userR.GetUserByEmail(ctx, email)
	if err != nil {
		u.logger.Info("recover confirm: unknown email", "email", email)
		return domain.ErrUserNotFound
	}

	stored, err := u.cache.Get(ctx, domain.AuthRecoveryCodeKey(email))
	if err != nil || stored != code {
		u.logger.Info("recover confirm: wrong recovery code", "email", email)
		return domain.ErrInvalidCode
	}

	hash, err := HashSecret(newSecret)
	if err != nil {
		u.logger.Error("failed to hash new secret", "error", err)
		return err
	}
	user.Password = hash
	if err := u.userR.UpdateUser(ctx, user); err != nil {
		u.logger.Error("failed to update user secret", "error", err, "user_id", user.UserID)
		return err
	}

	// Снятие блокировки: удаляем флаг блокировки, счётчик неудач и коды.
	_ = u.cache.Delete(ctx, domain.AuthLockKey(user.UserID))
	_ = u.cache.Delete(ctx, domain.AuthFailuresKey(user.UserID))
	_ = u.cache.Delete(ctx, domain.AuthCodeKey(user.UserID))
	_ = u.cache.Delete(ctx, domain.AuthRecoveryCodeKey(email))
	u.logger.Info("recover confirm: account unlocked, secret updated", "user_id", user.UserID)
	return nil
}
