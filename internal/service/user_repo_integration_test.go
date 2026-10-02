//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

func setupUserRepo(t *testing.T) (*pgxpool.Pool, domain.UserRepository) {
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
	pool, repo := setupUserRepo(t)
	ctx := context.Background()

	t.Run("StoreUser and GetUser", func(t *testing.T) {
		u := newUser("alice@example.com", 500, nil)
		if err := repo.StoreUser(ctx, u); err != nil {
			t.Fatalf("StoreUser: %v", err)
		}
		got, err := repo.GetUser(ctx, u.UserID)
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if got.Email != "alice@example.com" {
			t.Errorf("Email: got %q want alice@example.com", got.Email)
		}
		if got.Balance != 500 {
			t.Errorf("Balance: got %d want 500", got.Balance)
		}
	})

	t.Run("GetUser returns ErrUserNotFound for unknown ID", func(t *testing.T) {
		_, err := repo.GetUser(ctx, uuid.New())
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("UpdateUser modifies balance", func(t *testing.T) {
		u := newUser("bob@example.com", 1000, nil)
		if err := repo.StoreUser(ctx, u); err != nil {
			t.Fatalf("StoreUser: %v", err)
		}
		u.Balance = 800
		if err := repo.UpdateUser(ctx, u); err != nil {
			t.Fatalf("UpdateUser: %v", err)
		}
		got, err := repo.GetUser(ctx, u.UserID)
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if got.Balance != 800 {
			t.Errorf("Balance: got %d want 800", got.Balance)
		}
	})

	t.Run("GetUserByReferralCode", func(t *testing.T) {
		code := "REFCODE1"
		u := newUser("carol@example.com", 300, &code)
		if err := repo.StoreUser(ctx, u); err != nil {
			t.Fatalf("StoreUser: %v", err)
		}
		got, err := repo.GetUserByReferralCode(ctx, code)
		if err != nil {
			t.Fatalf("GetUserByReferralCode: %v", err)
		}
		if got.UserID != u.UserID {
			t.Errorf("UserID mismatch: got %v want %v", got.UserID, u.UserID)
		}
	})

	t.Run("StoreReferral links two users", func(t *testing.T) {
		// Триггер add_referral_promocodes выбирает случайный service_id,
		// поэтому в таблице services должна быть хотя бы одна запись.
		testdb.SeedService(t, pool, "Netflix")
		referrer := newUser("referrer@example.com", 100, nil)
		referred := newUser("referred@example.com", 100, nil)
		if err := repo.StoreUser(ctx, referrer); err != nil {
			t.Fatalf("StoreUser referrer: %v", err)
		}
		if err := repo.StoreUser(ctx, referred); err != nil {
			t.Fatalf("StoreUser referred: %v", err)
		}
		if err := repo.StoreReferral(ctx, referrer.UserID, referred.UserID); err != nil {
			t.Fatalf("StoreReferral: %v", err)
		}
	})

	t.Run("SetAppCurrentUserID succeeds", func(t *testing.T) {
		u := newUser("dave@example.com", 100, nil)
		if err := repo.StoreUser(ctx, u); err != nil {
			t.Fatalf("StoreUser: %v", err)
		}
		if err := repo.SetAppCurrentUserID(ctx, u.UserID); err != nil {
			t.Fatalf("SetAppCurrentUserID: %v", err)
		}
	})
}
