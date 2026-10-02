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
		pt.Tags("integration", "data-access", "users")

		pool, repo := setupUserRepo(pt)
		ctx := context.Background()

		pt.Run("StoreUser and GetUser", func(t provider.T) {
			u := newUser("alice@example.com", 500, nil)
			t.Require().NoError(repo.StoreUser(ctx, u))
			got, err := repo.GetUser(ctx, u.UserID)
			t.Require().NoError(err)
			t.Assert().Equal("alice@example.com", got.Email)
			t.Assert().Equal(500, got.Balance)
		})

		pt.Run("GetUser returns ErrUserNotFound for unknown ID", func(t provider.T) {
			_, err := repo.GetUser(ctx, uuid.New())
			t.Assert().Equal(domain.ErrUserNotFound, err)
		})

		pt.Run("UpdateUser modifies balance", func(t provider.T) {
			u := newUser("bob@example.com", 1000, nil)
			t.Require().NoError(repo.StoreUser(ctx, u))
			u.Balance = 800
			t.Require().NoError(repo.UpdateUser(ctx, u))
			got, err := repo.GetUser(ctx, u.UserID)
			t.Require().NoError(err)
			t.Assert().Equal(800, got.Balance)
		})

		pt.Run("GetUserByReferralCode", func(t provider.T) {
			code := "REFCODE1"
			u := newUser("carol@example.com", 300, &code)
			t.Require().NoError(repo.StoreUser(ctx, u))
			got, err := repo.GetUserByReferralCode(ctx, code)
			t.Require().NoError(err)
			t.Assert().Equal(u.UserID, got.UserID)
		})

		pt.Run("StoreReferral links two users", func(t provider.T) {
			// Триггер add_referral_promocodes выбирает случайный service_id,
			// поэтому в таблице services должна быть хотя бы одна запись.
			testdb.SeedService(t, pool, "Netflix")
			referrer := newUser("referrer@example.com", 100, nil)
			referred := newUser("referred@example.com", 100, nil)
			t.Require().NoError(repo.StoreUser(ctx, referrer))
			t.Require().NoError(repo.StoreUser(ctx, referred))
			t.Require().NoError(repo.StoreReferral(ctx, referrer.UserID, referred.UserID))
		})

		pt.Run("SetAppCurrentUserID succeeds", func(t provider.T) {
			u := newUser("dave@example.com", 100, nil)
			t.Require().NoError(repo.StoreUser(ctx, u))
			t.Require().NoError(repo.SetAppCurrentUserID(ctx, u.UserID))
		})
	})
}
