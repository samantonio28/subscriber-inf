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

func setupPaymentRepo(t testing.TB) (*pgxpool.Pool, domain.PaymentRepository, uuid.UUID) {
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
	runner.Run(t, "PaymentRepo integration", func(pt provider.T) {
		pt.Feature("PaymentRepo")
		pt.Tags("integration", "data-access", "payments")

		_, repo, user := setupPaymentRepo(pt)
		ctx := context.Background()

		pt.Run("StorePayment and GetUserPayments", func(t provider.T) {
			err := repo.StorePayment(ctx, domain.Payment{
				UserID:      user,
				CardNumber:  nil,
				Amount:      299,
				PaymentType: domain.PaymentEXPENCE,
			})
			t.Require().NoError(err)

			payments, err := repo.GetUserPayments(ctx, user)
			t.Require().NoError(err)
			t.Assert().Equal(1, len(payments))
			t.Assert().Equal(299, payments[0].Amount)
			t.Assert().Equal(domain.PaymentEXPENCE, payments[0].PaymentType)
		})
	})
}
