//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

func newPromocode(svc int, value string, discount, maxUses int) domain.Promocode {
	return domain.Promocode{
		ServiceID:    svc,
		Value:        value,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		Discount:     discount,
		MaxUses:      maxUses,
		CurUses:      0,
		Status:       domain.PromocodeStatusActive,
		DurationDays: 30,
	}
}

// TestPromocodeRepoIntegration проверяет доступ к данным промокодов против
// реального PostgreSQL. Каждый подтест независим (свой сброс и сидирование).
func TestPromocodeRepoIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	repo, err := NewPromocodeRepo(pool)
	if err != nil {
		t.Fatalf("NewPromocodeRepo: %v", err)
	}
	ctx := context.Background()

	seedSvc := func(t *testing.T) int {
		t.Helper()
		testdb.Reset(t, pool)
		return testdb.SeedService(t, pool, "Spotify")
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, newPromocode(svc, "SAVE20", 20, 5))
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero promocode id")
		}

		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Value != "SAVE20" {
			t.Errorf("Value: got %q want SAVE20", got.Value)
		}
		if got.Discount != 20 {
			t.Errorf("Discount: got %d want 20", got.Discount)
		}
	})

	t.Run("GetByCode", func(t *testing.T) {
		svc := seedSvc(t)
		if _, err := repo.Create(ctx, newPromocode(svc, "WELCOME5", 5, 10)); err != nil {
			t.Fatalf("Create: %v", err)
		}
		got, err := repo.GetByCode(ctx, "WELCOME5")
		if err != nil {
			t.Fatalf("GetByCode: %v", err)
		}
		if got.Value != "WELCOME5" {
			t.Errorf("Value: got %q want WELCOME5", got.Value)
		}
	})

	t.Run("GetByService returns promocodes of service", func(t *testing.T) {
		svc := seedSvc(t)
		if _, err := repo.Create(ctx, newPromocode(svc, "SAVE10", 10, 3)); err != nil {
			t.Fatalf("Create 1: %v", err)
		}
		if _, err := repo.Create(ctx, newPromocode(svc, "SAVE30", 30, 3)); err != nil {
			t.Fatalf("Create 2: %v", err)
		}
		list, err := repo.GetByService(ctx, svc)
		if err != nil {
			t.Fatalf("GetByService: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 promocodes, got %d", len(list))
		}
	})

	t.Run("Update modifies promocode", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, newPromocode(svc, "SAVE40", 40, 4))
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		pc := newPromocode(svc, "SAVE40", 50, 4)
		pc.PromocodeID = id
		if err := repo.Update(ctx, pc); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Discount != 50 {
			t.Errorf("Discount: got %d want 50", got.Discount)
		}
	})

	t.Run("IncrementUses increases cur_uses", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, newPromocode(svc, "SAVE60", 60, 5))
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if err := repo.IncrementUses(ctx, id); err != nil {
			t.Fatalf("IncrementUses: %v", err)
		}
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.CurUses != 1 {
			t.Errorf("CurUses: got %d want 1", got.CurUses)
		}
	})

	t.Run("Delete removes promocode", func(t *testing.T) {
		svc := seedSvc(t)
		id, err := repo.Create(ctx, newPromocode(svc, "SAVE70", 70, 1))
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
