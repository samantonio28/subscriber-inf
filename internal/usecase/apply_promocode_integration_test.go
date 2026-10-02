//go:build integration

package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestApplyPromocodeIntegration проверяет бизнес-логику применения промокода к
// активной подписке на реальной БД: пересчёт цены, обновление подписки и
// инкремент счётчика использований.
func TestApplyPromocodeIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	testdb.Reset(t, pool)

	subRepo, err := service.NewSubRepo(pool)
	if err != nil {
		t.Fatalf("NewSubRepo: %v", err)
	}
	promoRepo, err := service.NewPromocodeRepo(pool)
	if err != nil {
		t.Fatalf("NewPromocodeRepo: %v", err)
	}

	uc, err := NewApplyPromocodeUC(subRepo, promoRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("NewApplyPromocodeUC: %v", err)
	}

	svc := testdb.SeedService(t, pool, "Netflix")
	plan := testdb.SeedPlan(t, pool, svc, "Netflix Basic", 30, 1000)
	user := testdb.SeedUser(t, pool, "subscriber@example.com", 0, "user")

	start := testdb.FirstOfMonth(time.Now())
	end := testdb.FirstOfMonth(start.AddDate(0, 1, 0))
	subID := testdb.SeedSubscription(t, pool, user, plan, 1000, "usual", start, end)

	promoID := testdb.SeedPromocode(t, pool, svc, "SAVE20", 20, 5)

	out, err := uc.Apply(context.Background(), ApplyPromocodeInput{
		SubscriptionID: subID,
		PromocodeValue: "SAVE20",
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if out.DiscountApplied != 20 {
		t.Errorf("DiscountApplied: got %d want 20", out.DiscountApplied)
	}
	if out.NewPrice != 800 {
		t.Errorf("NewPrice: got %d want 800", out.NewPrice)
	}

	// Подписка должна быть обновлена: цена снижена, тип — promocode.
	sub, err := subRepo.Sub(context.Background(), domain.SubID(subID))
	if err != nil {
		t.Fatalf("Sub: %v", err)
	}
	if sub.Price != 800 {
		t.Errorf("sub Price: got %d want 800", sub.Price)
	}
	if sub.SubType.String() != "promocode" {
		t.Errorf("sub type: got %q want promocode", sub.SubType.String())
	}

	// Счётчик использований промокода должен увеличиться.
	promo, err := promoRepo.GetByID(context.Background(), domain.PromocodeID(promoID))
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if promo.CurUses != 1 {
		t.Errorf("promo CurUses: got %d want 1", promo.CurUses)
	}
}
