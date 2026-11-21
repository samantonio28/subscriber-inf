package usecase

// Юнит-тесты (london-стиль) юзкейсов аутентификации ЛР №5.
// Зависимости замоканы (MockUserRepository / MockAuthCache / MockMailer):
// БД и Redis не нужны — проверяется логика юзкейсов в изоляции.
// Каждый сабтест создаёт СВОИ моки (newAuthMocks(t) с t сабтеста), чтобы
// ожидания gomock не перетекали между сабтестами.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
	"go.uber.org/mock/gomock"
)

type authMocks struct {
	ctrl   *gomock.Controller
	userR  *mock.MockUserRepository
	cache  *mock.MockAuthCache
	mail   *mock.MockMailer
	logger *testutil.NopLogger
	ctx    context.Context
}

func newAuthMocks(t *testing.T) *authMocks {
	ctrl := gomock.NewController(t)
	return &authMocks{
		ctrl:   ctrl,
		userR:  mock.NewMockUserRepository(ctrl),
		cache:  mock.NewMockAuthCache(ctrl),
		mail:   mock.NewMockMailer(ctrl),
		logger: &testutil.NopLogger{},
		ctx:    context.Background(),
	}
}

func hashed(secret string) string {
	h, _ := HashSecret(secret)
	return h
}

// --- HashSecret / CheckSecret -------------------------------------------------

func TestHashAndCheckSecret(t *testing.T) {
	t.Run("bcrypt roundtrip", func(t *testing.T) {
		hash, err := HashSecret("s3cure-Pass")
		if err != nil {
			t.Fatalf("hash: %v", err)
		}
		if hash == "s3cure-Pass" {
			t.Error("hash must not equal plain secret")
		}
		if !strings.HasPrefix(hash, "$2") {
			t.Errorf("expected bcrypt hash, got %s", hash)
		}
		if !CheckSecret(hash, "s3cure-Pass") {
			t.Error("check must succeed for the right secret")
		}
		if CheckSecret(hash, "wrong-secret") {
			t.Error("check must fail for the wrong secret")
		}
	})
	t.Run("too short secret rejected", func(t *testing.T) {
		if _, err := HashSecret("123"); err == nil {
			t.Error("expected error for short secret")
		}
	})
}

// --- RegisterAuthUC ------------------------------------------------------------

func TestRegisterAuthUC(t *testing.T) {
	validDTO := RegisterAuthDTO{
		Email:    "tech.user@example.com",
		UserName: "Tech User",
		Age:      30,
		Secret:   "s3cure-Pass",
	}

	t.Run("constructor validation", func(t *testing.T) {
		m := newAuthMocks(t)
		if _, err := NewRegisterAuthUC(nil, m.logger); !errors.Is(err, domain.ErrInvalidUserRepo) {
			t.Errorf("expected ErrInvalidUserRepo, got %v", err)
		}
		if _, err := NewRegisterAuthUC(m.userR, nil); !errors.Is(err, domain.ErrInvalidLogger) {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
	})

	t.Run("successful registration stores bcrypt hash", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, err := NewRegisterAuthUC(m.userR, m.logger)
		if err != nil {
			t.Fatalf("usecase: %v", err)
		}
		m.userR.EXPECT().GetUserByEmail(m.ctx, validDTO.Email).Return(domain.User{}, domain.ErrUserNotFound)
		var stored domain.User
		m.userR.EXPECT().StoreUser(m.ctx, gomock.Any()).Do(func(_ context.Context, u domain.User) { stored = u }).Return(nil)

		userID, err := uc.Register(m.ctx, validDTO)
		if err != nil {
			t.Fatalf("register: %v", err)
		}
		if userID == uuid.Nil {
			t.Error("expected non-nil user id")
		}
		// Ключевое требование ЛР №5: в хранилище уходит ТОЛЬКО bcrypt-хэш.
		if stored.Password == validDTO.Secret {
			t.Error("plain secret must not be stored")
		}
		if !CheckSecret(stored.Password, validDTO.Secret) {
			t.Error("stored hash must match the secret")
		}
		if stored.Role != domain.RoleUser {
			t.Errorf("expected role user, got %s", stored.Role)
		}
	})

	t.Run("email already taken", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRegisterAuthUC(m.userR, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, validDTO.Email).Return(domain.User{UserID: uuid.New()}, nil)
		if _, err := uc.Register(m.ctx, validDTO); !errors.Is(err, domain.ErrEmailTaken) {
			t.Errorf("expected ErrEmailTaken, got %v", err)
		}
	})

	t.Run("weak secret rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRegisterAuthUC(m.userR, m.logger)
		if _, err := uc.Register(m.ctx, RegisterAuthDTO{Email: "x@y.z", UserName: "X", Secret: "123"}); !errors.Is(err, domain.ErrWeakSecret) {
			t.Errorf("expected ErrWeakSecret, got %v", err)
		}
	})

	t.Run("missing email rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRegisterAuthUC(m.userR, m.logger)
		if _, err := uc.Register(m.ctx, RegisterAuthDTO{UserName: "X", Secret: "s3cure-Pass"}); !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("repo error propagates", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRegisterAuthUC(m.userR, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, validDTO.Email).Return(domain.User{}, domain.ErrUserNotFound)
		m.userR.EXPECT().StoreUser(m.ctx, gomock.Any()).Return(errors.New("db down"))
		if _, err := uc.Register(m.ctx, validDTO); err == nil || errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("expected repo error, got %v", err)
		}
	})
}

// --- LoginUC: первый фактор + лимит попыток + блокировка -----------------------

func TestLoginUC(t *testing.T) {
	secret := "s3cure-Pass"
	user := domain.User{UserID: uuid.New(), Email: "tech.user@example.com", Password: hashed(secret)}

	t.Run("constructor validation", func(t *testing.T) {
		m := newAuthMocks(t)
		if _, err := NewLoginUC(nil, m.cache, m.mail, m.logger); !errors.Is(err, domain.ErrInvalidUserRepo) {
			t.Errorf("expected ErrInvalidUserRepo, got %v", err)
		}
		if _, err := NewLoginUC(m.userR, nil, m.mail, m.logger); !errors.Is(err, domain.ErrInvalidCache) {
			t.Errorf("expected ErrInvalidCache, got %v", err)
		}
		if _, err := NewLoginUC(m.userR, m.cache, nil, m.logger); err == nil {
			t.Error("expected error for nil mailer")
		}
	})

	t.Run("successful login sends 2FA code", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewLoginUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, user.Email).Return(user, nil)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(user.UserID)).Return(false, nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthFailuresKey(user.UserID)).Return(nil)
		m.cache.EXPECT().Set(m.ctx, domain.AuthCodeKey(user.UserID), gomock.Any(), domain.AuthCodeTTL).DoAndReturn(
			func(_ context.Context, _ string, code any, _ time.Duration) error {
				if len(code.(string)) != domain.AuthCodeDigits {
					t.Errorf("expected %d-digit code, got %v", domain.AuthCodeDigits, code)
				}
				return nil
			})
		m.mail.EXPECT().Send(m.ctx, user.Email, gomock.Any(), gomock.Any()).Return(nil)

		userID, err := uc.Login(m.ctx, LoginDTO{Email: user.Email, Secret: secret})
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if userID != user.UserID {
			t.Errorf("expected user id %s, got %s", user.UserID, userID)
		}
	})

	t.Run("unknown email looks like wrong credentials", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewLoginUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, "ghost@example.com").Return(domain.User{}, domain.ErrUserNotFound)
		if _, err := uc.Login(m.ctx, LoginDTO{Email: "ghost@example.com", Secret: "whatever"}); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong secret counted as failure", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewLoginUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, user.Email).Return(user, nil)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(user.UserID)).Return(false, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthFailuresKey(user.UserID)).Return("0", nil)
		m.cache.EXPECT().Set(m.ctx, domain.AuthFailuresKey(user.UserID), "1", domain.FailuresTTL).Return(nil)
		if _, err := uc.Login(m.ctx, LoginDTO{Email: user.Email, Secret: "nope-wrong"}); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("account locked after max attempts", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewLoginUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, user.Email).Return(user, nil)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(user.UserID)).Return(false, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthFailuresKey(user.UserID)).Return(fmt.Sprint(domain.MaxLoginAttempts - 1), nil)
		m.cache.EXPECT().Set(m.ctx, domain.AuthLockKey(user.UserID), "1", domain.LockDuration).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthFailuresKey(user.UserID)).Return(nil)
		if _, err := uc.Login(m.ctx, LoginDTO{Email: user.Email, Secret: "nope-wrong"}); !errors.Is(err, domain.ErrAccountLocked) {
			t.Errorf("expected ErrAccountLocked, got %v", err)
		}
	})

	t.Run("locked account rejects even correct secret", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewLoginUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, user.Email).Return(user, nil)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(user.UserID)).Return(true, nil)
		if _, err := uc.Login(m.ctx, LoginDTO{Email: user.Email, Secret: secret}); !errors.Is(err, domain.ErrAccountLocked) {
			t.Errorf("expected ErrAccountLocked, got %v", err)
		}
	})
}

// --- VerifyAuthCodeUC: второй фактор -------------------------------------------

func TestVerifyAuthCodeUC(t *testing.T) {
	userID := uuid.New()

	t.Run("successful verification is one-time", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewVerifyAuthCodeUC(m.cache, m.logger)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(userID)).Return(false, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthCodeKey(userID)).Return("123456", nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthCodeKey(userID)).Return(nil)
		if err := uc.Verify(m.ctx, VerifyCodeDTO{UserID: userID, Code: "123456"}); err != nil {
			t.Fatalf("verify: %v", err)
		}
	})

	t.Run("wrong code rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewVerifyAuthCodeUC(m.cache, m.logger)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(userID)).Return(false, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthCodeKey(userID)).Return("123456", nil)
		if err := uc.Verify(m.ctx, VerifyCodeDTO{UserID: userID, Code: "000000"}); !errors.Is(err, domain.ErrInvalidCode) {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
	})

	t.Run("expired code rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewVerifyAuthCodeUC(m.cache, m.logger)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(userID)).Return(false, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthCodeKey(userID)).Return("", errors.New("redis: nil"))
		if err := uc.Verify(m.ctx, VerifyCodeDTO{UserID: userID, Code: "123456"}); !errors.Is(err, domain.ErrInvalidCode) {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
	})

	t.Run("locked account cannot verify", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewVerifyAuthCodeUC(m.cache, m.logger)
		m.cache.EXPECT().Exists(m.ctx, domain.AuthLockKey(userID)).Return(true, nil)
		if err := uc.Verify(m.ctx, VerifyCodeDTO{UserID: userID, Code: "123456"}); !errors.Is(err, domain.ErrAccountLocked) {
			t.Errorf("expected ErrAccountLocked, got %v", err)
		}
	})

	t.Run("empty code rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewVerifyAuthCodeUC(m.cache, m.logger)
		if err := uc.Verify(m.ctx, VerifyCodeDTO{UserID: userID, Code: ""}); !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

// --- ChangePasswordUC: плановая смена пароля ------------------------------------

func TestChangePasswordUC(t *testing.T) {
	secret := "old-secret-1"
	user := domain.User{UserID: uuid.New(), Email: "tech.user@example.com", Password: hashed(secret), Role: domain.RoleUser}

	t.Run("successful change stores new hash and invalidates code", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewChangePasswordUC(m.userR, m.cache, m.logger)
		m.userR.EXPECT().GetUser(m.ctx, user.UserID).Return(user, nil)
		var updated domain.User
		m.userR.EXPECT().UpdateUser(m.ctx, gomock.Any()).Do(func(_ context.Context, u domain.User) { updated = u }).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthCodeKey(user.UserID)).Return(nil)

		if err := uc.Change(m.ctx, ChangePasswordDTO{UserID: user.UserID, CurrentSecret: secret, NewSecret: "new-secret-2"}); err != nil {
			t.Fatalf("change: %v", err)
		}
		if !CheckSecret(updated.Password, "new-secret-2") {
			t.Error("new hash must match the new secret")
		}
		if CheckSecret(updated.Password, secret) {
			t.Error("old secret must no longer match")
		}
	})

	t.Run("wrong current secret rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewChangePasswordUC(m.userR, m.cache, m.logger)
		m.userR.EXPECT().GetUser(m.ctx, user.UserID).Return(user, nil)
		if err := uc.Change(m.ctx, ChangePasswordDTO{UserID: user.UserID, CurrentSecret: "totally-wrong", NewSecret: "new-secret-2"}); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("weak new secret rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewChangePasswordUC(m.userR, m.cache, m.logger)
		if err := uc.Change(m.ctx, ChangePasswordDTO{UserID: user.UserID, CurrentSecret: secret, NewSecret: "123"}); !errors.Is(err, domain.ErrWeakSecret) {
			t.Errorf("expected ErrWeakSecret, got %v", err)
		}
	})

	t.Run("unknown user rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewChangePasswordUC(m.userR, m.cache, m.logger)
		ghost := uuid.New()
		m.userR.EXPECT().GetUser(m.ctx, ghost).Return(domain.User{}, domain.ErrUserNotFound)
		if err := uc.Change(m.ctx, ChangePasswordDTO{UserID: ghost, CurrentSecret: "x", NewSecret: "new-secret-2"}); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// --- RecoverAuthUC: восстановление при блокировке -------------------------------

func TestRecoverAuthUC(t *testing.T) {
	email := "tech.user@example.com"
	user := domain.User{UserID: uuid.New(), Email: email, Password: hashed("locked-secret"), Role: domain.RoleUser}

	t.Run("recover sends code to email", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRecoverAuthUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, email).Return(user, nil)
		m.cache.EXPECT().Set(m.ctx, domain.AuthRecoveryCodeKey(email), gomock.Any(), domain.AuthCodeTTL).Return(nil)
		m.mail.EXPECT().Send(m.ctx, email, gomock.Any(), gomock.Any()).Return(nil)
		if err := uc.Recover(m.ctx, email); err != nil {
			t.Fatalf("recover: %v", err)
		}
	})

	t.Run("recover unknown email", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRecoverAuthUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, "ghost@example.com").Return(domain.User{}, domain.ErrUserNotFound)
		if err := uc.Recover(m.ctx, "ghost@example.com"); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("confirm unlocks account and sets new secret", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRecoverAuthUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, email).Return(user, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthRecoveryCodeKey(email)).Return("654321", nil)
		var updated domain.User
		m.userR.EXPECT().UpdateUser(m.ctx, gomock.Any()).Do(func(_ context.Context, u domain.User) { updated = u }).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthLockKey(user.UserID)).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthFailuresKey(user.UserID)).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthCodeKey(user.UserID)).Return(nil)
		m.cache.EXPECT().Delete(m.ctx, domain.AuthRecoveryCodeKey(email)).Return(nil)

		if err := uc.Confirm(m.ctx, email, "654321", "fresh-secret-3"); err != nil {
			t.Fatalf("confirm: %v", err)
		}
		if !CheckSecret(updated.Password, "fresh-secret-3") {
			t.Error("new hash must match the new secret")
		}
	})

	t.Run("confirm with wrong code rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRecoverAuthUC(m.userR, m.cache, m.mail, m.logger)
		m.userR.EXPECT().GetUserByEmail(m.ctx, email).Return(user, nil)
		m.cache.EXPECT().Get(m.ctx, domain.AuthRecoveryCodeKey(email)).Return("654321", nil)
		if err := uc.Confirm(m.ctx, email, "000000", "fresh-secret-3"); !errors.Is(err, domain.ErrInvalidCode) {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
	})

	t.Run("confirm with weak secret rejected", func(t *testing.T) {
		m := newAuthMocks(t)
		uc, _ := NewRecoverAuthUC(m.userR, m.cache, m.mail, m.logger)
		if err := uc.Confirm(m.ctx, email, "654321", "123"); !errors.Is(err, domain.ErrWeakSecret) {
			t.Errorf("expected ErrWeakSecret, got %v", err)
		}
	})
}
