package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// Классические тесты: настоящие репозитории (UserRepo, SubRepo, PaymentRepo)
// поверх «замоканного PostgreSQL» (FakePgPool).

func TestPurchaseSubscriptionUC_Purchase_Classic(t *testing.T) {
	ctx := context.Background()
	userID := testutil.TestUUID()

	t.Run("successful purchase against mocked postgres", func(t *testing.T) {
		// Arrange: настоящие репозитории получают «замоканный коннект».
		fakePG := testutil.NewFakePgPool()

		// GetUser: SELECT ... FROM users WHERE user_id = $1
		fakePG.SetQueryRow("FROM users", testutil.NewStubRow(
			userID, "alice@example.com", "secret", "Alice", 25, 1000, "REF123", "user",
		))
		// StoreSub: INSERT INTO subscriptions ... RETURNING sub_id (в транзакции)
		fakePG.Tx().OnQueryRow("RETURNING sub_id", testutil.NewStubRow(42))
		// StorePayment: INSERT INTO payments ... RETURNING paym_id
		fakePG.SetQueryRow("RETURNING paym_id", testutil.NewStubRow(1))
		// UpdateUser: UPDATE users SET ...
		fakePG.SetExec("UPDATE users", testutil.CommandTagUpdate(1))

		userRepo, _ := service.NewUserRepo(fakePG)
		subRepo, _ := service.NewSubRepo(fakePG)
		paymentRepo, _ := service.NewPaymentRepo(fakePG)

		uc, err := NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})
		if err != nil {
			t.Fatalf("failed to create usecase: %v", err)
		}

		input := PurchaseSubscriptionDTO{
			UserID:       userID,
			ServiceName:  "Netflix",
			PlanID:       1,
			Price:        300,
			DurationDays: 30,
		}

		// Act
		id, err := uc.Purchase(ctx, input)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if id != 42 {
			t.Errorf("expected sub id 42, got %d", id)
		}
	})

	t.Run("insufficient balance rejected", func(t *testing.T) {
		// Arrange
		fakePG := testutil.NewFakePgPool()
		fakePG.SetQueryRow("FROM users", testutil.NewStubRow(
			userID, "bob@example.com", "secret", "Bob", 25, 100, "REF222", "user",
		))

		userRepo, _ := service.NewUserRepo(fakePG)
		subRepo, _ := service.NewSubRepo(fakePG)
		paymentRepo, _ := service.NewPaymentRepo(fakePG)

		uc, _ := NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})

		input := PurchaseSubscriptionDTO{
			UserID:       userID,
			ServiceName:  "Netflix",
			PlanID:       1,
			Price:        300,
			DurationDays: 30,
		}

		// Act
		id, err := uc.Purchase(ctx, input)

		// Assert
		if !errors.Is(err, domain.ErrInsufficientBalance) || id != 0 {
			t.Fatalf("expected ErrInsufficientBalance with id 0, got err=%v id=%d", err, id)
		}
	})
}

func TestTotalCostsUC_TotalCosts_Classic(t *testing.T) {
	ctx := context.Background()
	userID := testutil.TestUUID()

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	mar1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	// Arrange: «БД» содержит одну подписку Netflix на 1 месяц (январь).
	fakePG := testutil.NewFakePgPool()
	fakePG.SetQuery("SELECT sub_id FROM subscriptions", testutil.NewStubRows([]any{1}))
	fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 100, "usual", jan1, feb1))

	subRepo, _ := service.NewSubRepo(fakePG)
	uc, err := NewTotalCostsUC(subRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	input := SubsFilterDTO{
		StartDate:   jan1,
		EndDate:     mar1,
		UserID:      userID,
		ServiceName: "Netflix",
		SubType:     "usual",
	}

	// Act
	sum, subs, err := uc.TotalCosts(ctx, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription in result, got %d", len(subs))
	}
	// Цена 100 × 1 месяц (январь) = 100.
	if sum != 100 {
		t.Errorf("expected total cost 100, got %d", sum)
	}
}
