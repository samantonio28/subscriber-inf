//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestSubRepoIntegration проверяет компонент доступа к данным (SubRepo) против
// реального PostgreSQL: полный жизненный цикл подписки и расчёт стоимости.
//
// Каждый подтест сам сбрасывает БД и сидирует нужные сущности, поэтому
// подтесты независимы друг от друга (устойчивы к -shuffle, повторному запуску
// и запуску по отдельности через -run).
func TestSubRepoIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	repo, err := NewSubRepo(pool)
	if err != nil {
		t.Fatalf("NewSubRepo: %v", err)
	}
	ctx := context.Background()

	// seedSubContext очищает БД и создаёт service → plan → user, возвращает их ID.
	seedSubContext := func(t *testing.T) (int, uuid.UUID) {
		t.Helper()
		testdb.Reset(t, pool)
		svc := testdb.SeedService(t, pool, "Netflix")
		plan := testdb.SeedPlan(t, pool, svc, "Netflix Basic", 30, 299)
		user := testdb.SeedUser(t, pool, "user@example.com", 1000, "user")
		return plan, user
	}

	start := testdb.FirstOfMonth(time.Now())
	end := testdb.FirstOfMonth(start.AddDate(0, 1, 0))

	t.Run("StoreSub and retrieve by ID", func(t *testing.T) {
		plan, user := seedSubContext(t)
		sub := domain.Subscription{
			UserID:      user,
			ServiceName: "Netflix", // имя выводится через join plan → service
			Price:       299,
			SubType:     domain.SubTypeUsual,
			StartDate:   start,
			EndDate:     end,
			PlanID:      plan,
		}

		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Fatalf("StoreSub: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero sub id")
		}

		got, err := repo.Sub(ctx, id)
		if err != nil {
			t.Fatalf("Sub: %v", err)
		}
		if got.UserID != user {
			t.Errorf("UserID: got %v want %v", got.UserID, user)
		}
		if got.Price != 299 {
			t.Errorf("Price: got %d want 299", got.Price)
		}
		if got.ServiceName != "Netflix" {
			t.Errorf("ServiceName: got %q want Netflix", got.ServiceName)
		}
		if got.PlanID != plan {
			t.Errorf("PlanID: got %d want %d", got.PlanID, plan)
		}
	})

	t.Run("UserSubs returns subscriptions for user", func(t *testing.T) {
		plan, user := seedSubContext(t)
		sub1 := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
		sub2 := domain.Subscription{UserID: user, Price: 200, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
		if _, err := repo.StoreSub(ctx, sub1); err != nil {
			t.Fatalf("StoreSub 1: %v", err)
		}
		if _, err := repo.StoreSub(ctx, sub2); err != nil {
			t.Fatalf("StoreSub 2: %v", err)
		}

		subs, err := repo.UserSubs(ctx, user)
		if err != nil {
			t.Fatalf("UserSubs: %v", err)
		}
		if len(subs) != 2 {
			t.Fatalf("expected 2 subscriptions, got %d", len(subs))
		}
	})

	t.Run("UpdateSub modifies existing subscription", func(t *testing.T) {
		plan, user := seedSubContext(t)
		sub := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Fatalf("StoreSub: %v", err)
		}

		updated := domain.Subscription{
			SubId:       id,
			UserID:      user,
			Price:       399,
			SubType:     domain.SubTypePromocode,
			StartDate:   start,
			EndDate:     end,
			PlanID:      plan,
		}
		if err := repo.UpdateSub(ctx, updated); err != nil {
			t.Fatalf("UpdateSub: %v", err)
		}

		got, err := repo.Sub(ctx, id)
		if err != nil {
			t.Fatalf("Sub: %v", err)
		}
		if got.Price != 399 {
			t.Errorf("Price: got %d want 399", got.Price)
		}
		if got.SubType.String() != "promocode" {
			t.Errorf("SubType: got %q want promocode", got.SubType.String())
		}
	})

	t.Run("DeleteSub removes subscription", func(t *testing.T) {
		plan, user := seedSubContext(t)
		sub := domain.Subscription{UserID: user, Price: 50, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Fatalf("StoreSub: %v", err)
		}

		if err := repo.DeleteSub(ctx, id); err != nil {
			t.Fatalf("DeleteSub: %v", err)
		}
		if _, err := repo.Sub(ctx, id); err == nil {
			t.Fatal("expected error after deletion, got nil")
		}
	})

	t.Run("Sub returns error for non-existent ID", func(t *testing.T) {
		seedSubContext(t)
		if _, err := repo.Sub(ctx, domain.SubID(999999)); err == nil {
			t.Fatal("expected error for non-existent ID, got nil")
		}
	})

	t.Run("SubsTotalCosts computes cost for period", func(t *testing.T) {
		plan, user := seedSubContext(t)
		jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
		sub := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: jan, EndDate: feb, PlanID: plan}
		if _, err := repo.StoreSub(ctx, sub); err != nil {
			t.Fatalf("StoreSub: %v", err)
		}

		sum, subs, err := repo.SubsTotalCosts(ctx, domain.SubsFilter{
			StartDate:   jan,
			EndDate:     mar,
			UserID:      user,
			ServiceName: "Netflix",
			SubType:     domain.SubTypeUsual,
		})
		if err != nil {
			t.Fatalf("SubsTotalCosts: %v", err)
		}
		if sum != 100 {
			t.Errorf("sum: got %d want 100", sum)
		}
		if len(subs) != 1 {
			t.Errorf("subs len: got %d want 1", len(subs))
		}
	})
}
