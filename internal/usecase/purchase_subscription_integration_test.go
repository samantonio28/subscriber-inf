//go:build integration

package usecase

import (
	"context"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestPurchaseSubscriptionIntegration проверяет бизнес-логику покупки подписки
// на реальной БД: создание подписки, списание средств и запись платежа.
func TestPurchaseSubscriptionIntegration(t *testing.T) {
	runner.Run(t, "PurchaseSubscription integration", func(pt provider.T) {
		pt.Feature("PurchaseSubscription")
		pt.Tags("integration", "business-logic", "payments")

		pool := testdb.Connect(pt)
		testdb.Reset(pt, pool)

		subRepo, err := service.NewSubRepo(pool)
		pt.Require().NoError(err)
		userRepo, err := service.NewUserRepo(pool)
		pt.Require().NoError(err)
		paymentRepo, err := service.NewPaymentRepo(pool)
		pt.Require().NoError(err)

		uc, err := NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})
		pt.Require().NoError(err)

		svc := testdb.SeedService(pt, pool, "Netflix")
		plan := testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 299)
		user := testdb.SeedUser(pt, pool, "buyer@example.com", 1000, "user")

		subID, err := uc.Purchase(context.Background(), PurchaseSubscriptionDTO{
			UserID:       user,
			ServiceName:  "Netflix",
			PlanID:       plan,
			Price:        299,
			DurationDays: 30,
		})
		pt.Require().NoError(err)
		pt.Assert().NotZero(int(subID))

		// Баланс пользователя должен уменьшиться на цену подписки.
		gotUser, err := userRepo.GetUser(context.Background(), user)
		pt.Require().NoError(err)
		pt.Assert().Equal(1000-299, gotUser.Balance)

		// Должен быть записан ровно один платёж-расход.
		payments, err := paymentRepo.GetUserPayments(context.Background(), user)
		pt.Require().NoError(err)
		pt.Assert().Equal(1, len(payments))
		pt.Assert().Equal(299, payments[0].Amount)
		pt.Assert().Equal(domain.PaymentEXPENCE, payments[0].PaymentType)
	})
}
