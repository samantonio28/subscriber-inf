//go:build integration

package usecase

import (
	"context"
	"testing"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestPurchaseSubscriptionIntegration проверяет бизнес-логику покупки подписки
// на реальной БД: создание подписки, списание средств и запись платежа.
func TestPurchaseSubscriptionIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	testdb.Reset(t, pool)

	subRepo, err := service.NewSubRepo(pool)
	if err != nil {
		t.Fatalf("NewSubRepo: %v", err)
	}
	userRepo, err := service.NewUserRepo(pool)
	if err != nil {
		t.Fatalf("NewUserRepo: %v", err)
	}
	paymentRepo, err := service.NewPaymentRepo(pool)
	if err != nil {
		t.Fatalf("NewPaymentRepo: %v", err)
	}

	uc, err := NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("NewPurchaseSubscriptionUC: %v", err)
	}

	svc := testdb.SeedService(t, pool, "Netflix")
	plan := testdb.SeedPlan(t, pool, svc, "Netflix Basic", 30, 299)
	user := testdb.SeedUser(t, pool, "buyer@example.com", 1000, "user")

	subID, err := uc.Purchase(context.Background(), PurchaseSubscriptionDTO{
		UserID:       user,
		ServiceName:  "Netflix",
		PlanID:       plan,
		Price:        299,
		DurationDays: 30,
	})
	if err != nil {
		t.Fatalf("Purchase: %v", err)
	}
	if subID == 0 {
		t.Fatal("expected non-zero sub id")
	}

	// Баланс пользователя должен уменьшиться на цену подписки.
	gotUser, err := userRepo.GetUser(context.Background(), user)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if gotUser.Balance != 1000-299 {
		t.Errorf("Balance: got %d want %d", gotUser.Balance, 1000-299)
	}

	// Должен быть записан ровно один платёж-расход.
	payments, err := paymentRepo.GetUserPayments(context.Background(), user)
	if err != nil {
		t.Fatalf("GetUserPayments: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(payments))
	}
	if payments[0].Amount != 299 {
		t.Errorf("payment Amount: got %d want 299", payments[0].Amount)
	}
	if payments[0].PaymentType != domain.PaymentEXPENCE {
		t.Errorf("payment type: got %q want expence", payments[0].PaymentType)
	}
}
