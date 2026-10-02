//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

func setupUserRepo(t testing.TB) (*pgxpool.Pool, domain.UserRepository) {
	t.Helper()
	pool := testdb.Connect(t)
	testdb.Reset(t, pool)

	repo, err := NewUserRepo(pool)
	if err != nil {
		t.Fatalf("NewUserRepo: %v", err)
	}
	return pool, repo
}

func newUser(email string, balance int, referralCode *string) domain.User {
	return domain.User{
		UserID:       uuid.New(),
		Email:        email,
		Password:     "secret",
		UserName:     "Test User",
		Age:          25,
		Balance:      balance,
		ReferralCode: referralCode,
		Role:         domain.RoleUser,
	}
}

// TestUserRepoIntegration проверяет доступ к данным пользователей против
// реального PostgreSQL, включая реферальную связь (с триггером промокодов).
func TestUserRepoIntegration(t *testing.T) {
	runner.Run(t, "UserRepo integration", func(pt provider.T) {
		pt.Feature("UserRepo")
		pt.Description("Доступ к данным пользователей (UserRepo) против реального PostgreSQL: " +
			"создание, чтение, обновление баланса, поиск по реферальному коду, реферальные связи и RLS-параметр сессии.")

		pool, repo := setupUserRepo(pt)
		ctx := context.Background()

		pt.Run("StoreUser and GetUser", func(t provider.T) {
			t.Description("Создание пользователя и его чтение по ID.")
			t.WithNewStep("Создание пользователя alice с балансом 500", func(s provider.StepCtx) {
				u := newUser("alice@example.com", 500, nil)
				s.Require().NoError(repo.StoreUser(ctx, u), "StoreUser не должен вернуть ошибку")

				got, err := repo.GetUser(ctx, u.UserID)
				s.Require().NoError(err, "GetUser не должен вернуть ошибку")
				s.Assert().Equal("alice@example.com", got.Email, "email должен совпадать")
				s.Assert().Equal(500, got.Balance, "баланс должен быть 500")
			})
		})

		pt.Run("GetUser returns ErrUserNotFound for unknown ID", func(t provider.T) {
			t.Description("Чтение несуществующего пользователя должно вернуть ErrUserNotFound.")
			t.WithNewStep("Чтение по случайному UUID", func(s provider.StepCtx) {
				_, err := repo.GetUser(ctx, uuid.New())
				s.Assert().Equal(domain.ErrUserNotFound, err, "должна вернуться ErrUserNotFound")
			})
		})

		pt.Run("UpdateUser modifies balance", func(t provider.T) {
			t.Description("Обновление баланса пользователя с 1000 до 800.")
			t.WithNewStep("Создание bob и обновление баланса", func(s provider.StepCtx) {
				u := newUser("bob@example.com", 1000, nil)
				s.Require().NoError(repo.StoreUser(ctx, u), "пользователь должен создаться")
				u.Balance = 800
				s.Require().NoError(repo.UpdateUser(ctx, u), "UpdateUser не должен вернуть ошибку")

				got, err := repo.GetUser(ctx, u.UserID)
				s.Require().NoError(err, "пользователь должен читаться после обновления")
				s.Assert().Equal(800, got.Balance, "баланс должен стать 800")
			})
		})

		pt.Run("GetUserByReferralCode", func(t provider.T) {
			t.Description("Поиск пользователя по реферальному коду.")
			t.WithNewStep("Создание carol с кодом REFCODE1 и поиск", func(s provider.StepCtx) {
				code := "REFCODE1"
				u := newUser("carol@example.com", 300, &code)
				s.Require().NoError(repo.StoreUser(ctx, u), "пользователь должен создаться")

				got, err := repo.GetUserByReferralCode(ctx, code)
				s.Require().NoError(err, "GetUserByReferralCode не должен вернуть ошибку")
				s.Assert().Equal(u.UserID, got.UserID, "должен вернуться тот же пользователь")
			})
		})

		pt.Run("StoreReferral links two users", func(t provider.T) {
			t.Description("Создание реферальной связи между двумя пользователями (триггер добавляет промокоды).")
			t.WithNewStep("Создание связи referrer → referred", func(s provider.StepCtx) {
				testdb.SeedService(t, pool, "Netflix")
				referrer := newUser("referrer@example.com", 100, nil)
				referred := newUser("referred@example.com", 100, nil)
				s.Require().NoError(repo.StoreUser(ctx, referrer), "referrer должен создаться")
				s.Require().NoError(repo.StoreUser(ctx, referred), "referred должен создаться")
				s.Require().NoError(repo.StoreReferral(ctx, referrer.UserID, referred.UserID), "StoreReferral не должен вернуть ошибку")
			})
		})

		pt.Run("SetAppCurrentUserID succeeds", func(t provider.T) {
			t.Description("Установка параметра сессии app.current_user_id для RLS.")
			t.WithNewStep("Создание dave и установка параметра сессии", func(s provider.StepCtx) {
				u := newUser("dave@example.com", 100, nil)
				s.Require().NoError(repo.StoreUser(ctx, u), "пользователь должен создаться")
				s.Require().NoError(repo.SetAppCurrentUserID(ctx, u.UserID), "SetAppCurrentUserID не должен вернуть ошибку")
			})
		})
	})
}
