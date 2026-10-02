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

func setupPaymentRepo(t *testing.T) (*pgxpool.Pool, domain.PaymentRepository, uuid.UUID) {
	t.Helper()
	pool := testdb.Connect(t)
	testdb.Reset(t, pool)

	repo, err := NewPaymentRepo(pool)
	if err != nil {
		t.Fatalf("NewPaymentRepo: %v", err)
	}

	user := testdb.SeedUser(t, pool, "payer@example.com", 1000, "user")
	return pool, repo, user
}

// TestPaymentRepoIntegration проверяет запись и чтение платежей против
// реального PostgreSQL (платёж-расход с card_number = NULL).
func TestPaymentRepoIntegration(t *testing.T) {
	_, repo, user := setupPaymentRepo(t)
	ctx := context.Background()

	t.Run("StorePayment and GetUserPayments", func(t *testing.T) {
		err := repo.StorePayment(ctx, domain.Payment{
			UserID:      user,
			CardNumber:  nil,
			Amount:      299,
			PaymentType: domain.PaymentEXPENCE,
		})
		if err != nil {
			t.Fatalf("StorePayment: %v", err)
		}

		payments, err := repo.GetUserPayments(ctx, user)
		if err != nil {
			t.Fatalf("GetUserPayments: %v", err)
		}
		if len(payments) != 1 {
			t.Fatalf("expected 1 payment, got %d", len(payments))
		}
		if payments[0].Amount != 299 {
			t.Errorf("Amount: got %d want 299", payments[0].Amount)
		}
		if payments[0].PaymentType != domain.PaymentEXPENCE {
			t.Errorf("PaymentType: got %q want expence", payments[0].PaymentType)
		}
	})
}
