package domain

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../mocks/auth_cache.go -package=mocks . AuthCache
//go:generate mockgen -destination=../mocks/mailer.go -package=mocks . Mailer

// Ошибки аутентификации. Каждая ошибка мапится на конкретный HTTP-статус в
// delivery: ErrInvalidCredentials → 401, ErrAccountLocked → 423, ErrInvalidCode → 401.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked: too many failed login attempts")
	ErrInvalidCode        = errors.New("invalid or expired confirmation code")
	ErrEmailTaken         = errors.New("email already registered")
	ErrWeakSecret         = errors.New("login secret is too short (minimum 6 characters)")
)

// Политика аутентификации (ЛР №5).
const (
	// MaxLoginAttempts — сколько раз подряд можно ввести неверный пароль до блокировки.
	MaxLoginAttempts = 3
	// LockDuration — на сколько блокируется аккаунт после превышения лимита попыток.
	LockDuration = 5 * time.Minute
	// AuthCodeTTL — срок жизни одноразового кода подтверждения (2FA) и кода восстановления.
	AuthCodeTTL = 5 * time.Minute
	// FailuresTTL — срок жизни счётчика неудачных попыток входа.
	FailuresTTL = 15 * time.Minute
	// MailOutboxTTL — срок жизни «письма» в почтовом ящике (Redis outbox).
	MailOutboxTTL = 10 * time.Minute
	// AuthCodeDigits — длина одноразового кода.
	AuthCodeDigits = 6
	// MinSecretLen — минимальная длина пароля.
	MinSecretLen = 6
)

// AuthCache — ключ-значение хранилище состояния аутентификации (Redis):
// счётчик неудачных попыток, флаг блокировки, одноразовые коды.
// Интерфейс совпадает с методами internal/redis.Client, поэтому клиент Redis
// реализует его напрямую, а в unit-тестах используется мок.
type AuthCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// Mailer — отправка писем (второй фактор — email). Продакшн-реализация — SMTP;
// тестовая/dev-реализация — outbox в Redis (service.OutboxMailer), который
// E2E-тесты читают как «почтовый ящик» пользователя.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Ключи состояния аутентификации в Redis.
const (
	authFailuresPrefix = "auth:failures:"
	authLockPrefix     = "auth:lock:"
	authCodePrefix     = "auth:code:"
	authRecoveryPrefix = "auth:recovery:"
	mailOutboxPrefix   = "mail:outbox:"
)

// AuthFailuresKey — счётчик подряд идущих неудачных попыток входа.
func AuthFailuresKey(userID uuid.UUID) string {
	return authFailuresPrefix + userID.String()
}

// AuthLockKey — признак блокировки аккаунта.
func AuthLockKey(userID uuid.UUID) string {
	return authLockPrefix + userID.String()
}

// AuthCodeKey — одноразовый код 2FA текущей сессии входа.
func AuthCodeKey(userID uuid.UUID) string {
	return authCodePrefix + userID.String()
}

// AuthRecoveryCodeKey — код восстановления (снимает блокировку).
func AuthRecoveryCodeKey(email string) string {
	return authRecoveryPrefix + email
}

// MailOutboxKey — «почтовый ящик»: последнее письмо, отправленное на email.
func MailOutboxKey(email string) string {
	return mailOutboxPrefix + email
}

// GenerateAuthCode генерирует одноразовый цифровой код (crypto/rand).
func GenerateAuthCode() (string, error) {
	const digits = "0123456789"
	buf := make([]byte, AuthCodeDigits)
	if err := fillRandom(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = digits[int(b)%len(digits)]
	}
	return string(buf), nil
}

// SendAuthCodeLetter формирует тело письма с одноразовым кодом.
func SendAuthCodeLetter(code string) string {
	return fmt.Sprintf("Ваш код подтверждения: %s. Код действует %d минут.", code, int(AuthCodeTTL.Minutes()))
}

// SendRecoveryCodeLetter формирует тело письма с кодом восстановления.
func SendRecoveryCodeLetter(code string) string {
	return fmt.Sprintf("Ваш код восстановления доступа: %s. Код действует %d минут.", code, int(AuthCodeTTL.Minutes()))
}

func fillRandom(buf []byte) error {
	_, err := rand.Read(buf)
	return err
}
