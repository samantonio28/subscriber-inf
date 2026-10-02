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
		pt.Description("Доступ к данным платежей (PaymentRepo) против реального PostgreSQL: " +
			"запись платежа-расхода и его чтение по пользователю.")

		_, repo, user := setupPaymentRepo(pt)
		ctx := context.Background()

		pt.Run("StorePayment and GetUserPayments", func(t provider.T) {
			t.Description("Запись платежа-расхода на 299 и его чтение по пользователю.")
			t.WithNewStep("Запись платежа-расхода (card_number = NULL)", func(s provider.StepCtx) {
				err := repo.StorePayment(ctx, domain.Payment{
					UserID: user, CardNumber: nil, Amount: 299, PaymentType: domain.PaymentEXPENCE,
				})
				s.Require().NoError(err, "StorePayment не должен вернуть ошибку")
			})
			t.WithNewStep("Чтение платежей пользователя", func(s provider.StepCtx) {
				payments, err := repo.GetUserPayments(ctx, user)
				s.Require().NoError(err, "GetUserPayments не должен вернуть ошибку")
				s.Assert().Equal(1, len(payments), "должен быть ровно один платёж")
				s.Assert().Equal(299, payments[0].Amount, "сумма платежа должна быть 299")
				s.Assert().Equal(domain.PaymentEXPENCE, payments[0].PaymentType, "тип платежа должен быть expence")
			})
		})
	})
}
