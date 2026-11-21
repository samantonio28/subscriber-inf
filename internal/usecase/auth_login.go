package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

// LoginUC — первый фактор аутентификации (email + пароль) и запуск второго:
// при успешной проверке пароля генерируется одноразовый код, который
// отправляется пользователю на email.
//
// Ограничение попыток ввода: после MaxLoginAttempts подряд неверных паролей
// аккаунт блокируется на LockDuration (флаг в Redis). Пароли сравниваются
// с bcrypt-хэшем, в БД открытых паролей нет.
type LoginUC struct {
	userR  domain.UserRepository
	cache  domain.AuthCache
	mail   domain.Mailer
	logger logger.Logger
}

func NewLoginUC(userR domain.UserRepository, cache domain.AuthCache, mail domain.Mailer, logger logger.Logger) (*LoginUC, error) {
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
	return &LoginUC{userR: userR, cache: cache, mail: mail, logger: logger}, nil
}

type LoginDTO struct {
	Email  string
	Secret string
}

// Login проверяет первый фактор и отправляет код 2FA на email.
// Ошибки: ErrInvalidCredentials (первый фактор не пройден), ErrAccountLocked
// (лимит попыток превышен), ErrUserNotFound мапится в ErrInvalidCredentials,
// чтобы не раскрывать наличие аккаунта.
func (u *LoginUC) Login(ctx context.Context, dto LoginDTO) (uuid.UUID, error) {
	if dto.Email == "" || dto.Secret == "" {
		return uuid.Nil, fmt.Errorf("%w: email and login secret are required", domain.ErrInvalidInput)
	}

	user, err := u.userR.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		// Не раскрываем, существует ли аккаунт: и «нет такого email», и
		// «неверный пароль» выглядят одинаково (401).
		u.logger.Info("login failed: unknown email", "email", dto.Email)
		return uuid.Nil, domain.ErrInvalidCredentials
	}

	// Блокировка: если аккаунт залочен, вход невозможен до истечения срока
	// или до подтверждения восстановления.
	locked, err := u.cache.Exists(ctx, domain.AuthLockKey(user.UserID))
	if err != nil {
		u.logger.Error("failed to check lock state", "error", err, "user_id", user.UserID)
		return uuid.Nil, fmt.Errorf("check lock state: %w", err)
	}
	if locked {
		u.logger.Info("login rejected: account locked", "user_id", user.UserID)
		return uuid.Nil, domain.ErrAccountLocked
	}

	// Первый фактор: пароль.
	if !CheckSecret(user.Password, dto.Secret) {
		return uuid.Nil, u.registerFailedAttempt(ctx, user)
	}

	// Успешный первый фактор: сбрасываем счётчик неудач и отправляем код 2FA.
	if err := u.cache.Delete(ctx, domain.AuthFailuresKey(user.UserID)); err != nil {
		u.logger.Warn("failed to reset failures counter", "error", err, "user_id", user.UserID)
	}
	if err := u.sendAuthCode(ctx, user); err != nil {
		return uuid.Nil, err
	}
	u.logger.Info("login: first factor ok, code sent", "user_id", user.UserID)
	return user.UserID, nil
}

// registerFailedAttempt инкрементирует счётчик подряд идущих неудачных
// попыток входа и блокирует аккаунт при превышении MaxLoginAttempts.
func (u *LoginUC) registerFailedAttempt(ctx context.Context, user domain.User) error {
	failures := 0
	if raw, err := u.cache.Get(ctx, domain.AuthFailuresKey(user.UserID)); err == nil {
		if n, parseErr := strconv.Atoi(raw); parseErr == nil {
			failures = n
		}
	}
	failures++
	if failures >= domain.MaxLoginAttempts {
		if err := u.cache.Set(ctx, domain.AuthLockKey(user.UserID), "1", domain.LockDuration); err != nil {
			u.logger.Error("failed to lock account", "error", err, "user_id", user.UserID)
			return fmt.Errorf("lock account: %w", err)
		}
		_ = u.cache.Delete(ctx, domain.AuthFailuresKey(user.UserID))
		u.logger.Warn("account locked after too many failed attempts", "user_id", user.UserID, "attempts", failures)
		return domain.ErrAccountLocked
	}
	if err := u.cache.Set(ctx, domain.AuthFailuresKey(user.UserID), strconv.Itoa(failures), domain.FailuresTTL); err != nil {
		u.logger.Error("failed to store failures counter", "error", err, "user_id", user.UserID)
	}
	u.logger.Info("login failed: wrong secret", "user_id", user.UserID, "attempt", failures)
	return domain.ErrInvalidCredentials
}

// sendAuthCode генерирует одноразовый код, сохраняет его в Redis (TTL AuthCodeTTL)
// и отправляет пользователю «на email» через Mailer.
func (u *LoginUC) sendAuthCode(ctx context.Context, user domain.User) error {
	code, err := domain.GenerateAuthCode()
	if err != nil {
		u.logger.Error("failed to generate auth code", "error", err)
		return fmt.Errorf("generate auth code: %w", err)
	}
	if err := u.cache.Set(ctx, domain.AuthCodeKey(user.UserID), code, domain.AuthCodeTTL); err != nil {
		u.logger.Error("failed to store auth code", "error", err, "user_id", user.UserID)
		return fmt.Errorf("store auth code: %w", err)
	}
	if err := u.mail.Send(ctx, user.Email, "Код подтверждения входа", domain.SendAuthCodeLetter(code)); err != nil {
		u.logger.Error("failed to send auth code", "error", err, "user_id", user.UserID)
		return fmt.Errorf("send auth code: %w", err)
	}
	return nil
}
