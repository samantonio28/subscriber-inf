//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestSubscriptionPlanRepoIntegration проверяет доступ к данным планов подписок
// против реального PostgreSQL. Каждый подтест независим.
func TestSubscriptionPlanRepoIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	repo, err := NewSubscriptionPlanRepo(pool)
	if err != nil {
		t.Fatalf("NewSubscriptionPlanRepo: %v", err)
	}
	ctx := context.Background()

	seedSvc := func(t *testing.T) int {
		t.Helper()
		testdb.Reset(t, pool)
		return testdb.SeedService(t, pool, "YouTube")
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero plan id")
		}

		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "YouTube Premium" {
			t.Errorf("Name: got %q want 'YouTube Premium'", got.Name)
		}
		if got.Price != 499 {
			t.Errorf("Price: got %d want 499", got.Price)
		}
	})

	t.Run("GetByService returns plans of service", func(t *testing.T) {
		svc := seedSvc(t)
		if _, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Family", DurationDays: 30, Price: 899}); err != nil {
			t.Fatalf("Create: %v", err)
		}
		list, err := repo.GetByService(ctx, svc)
		if err != nil {
			t.Fatalf("GetByService: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 plan, got %d", len(list))
		}
	})

	t.Run("GetAll returns all plans", func(t *testing.T) {
		svc := seedSvc(t)
		if _, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499}); err != nil {
			t.Fatalf("Create: %v", err)
		}
		list, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 plan, got %d", len(list))
		}
	})

	t.Run("Update modifies plan", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 299})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		err = repo.Update(ctx, domain.SubscriptionPlan{PlanID: id, ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 349})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Price != 349 {
			t.Errorf("Price: got %d want 349", got.Price)
		}
	})

	t.Run("Delete removes plan", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Music", DurationDays: 30, Price: 199})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if err := repo.Delete(ctx, id); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, id); err == nil {
			t.Fatal("expected error after deletion, got nil")
		}
	})
}
