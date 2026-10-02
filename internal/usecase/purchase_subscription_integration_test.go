//go:build integration

package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
		pt.Description("Бизнес-логика покупки подписки на реальной БД: при успешной покупке создаётся подписка, " +
			"списывается баланс пользователя и записывается платёж-расход.")

		var (
			userRepo    domain.UserRepository
			paymentRepo domain.PaymentRepository
			uc          *PurchaseSubscriptionUC
			user        uuid.UUID
			plan        int
		)

		pool := testdb.Connect(pt)
		testdb.Reset(pt, pool)

		pt.WithNewStep("Сборка репозиториев и usecase", func(s provider.StepCtx) {
			var err error
			subRepo, err := service.NewSubRepo(pool)
			s.Require().NoError(err, "SubRepo должен создаться")
			userRepo, err = service.NewUserRepo(pool)
			s.Require().NoError(err, "UserRepo должен создаться")
			paymentRepo, err = service.NewPaymentRepo(pool)
			s.Require().NoError(err, "PaymentRepo должен создаться")
			uc, err = NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})
			s.Require().NoError(err, "PurchaseSubscriptionUC должен создаться")
		})

		pt.WithNewStep("Подготовка данных: сервис, план, пользователь с балансом 1000", func(s provider.StepCtx) {
			svc := testdb.SeedService(pt, pool, "Netflix")
			plan = testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 299)
			user = testdb.SeedUser(pt, pool, "buyer@example.com", 1000, "user")
			s.Assert().NotEqual(uuid.Nil, user, "пользователь должен создаться с ненулевым UUID")
		})

		pt.WithNewStep("Покупка подписки за 299", func(s provider.StepCtx) {
			subID, err := uc.Purchase(context.Background(), PurchaseSubscriptionDTO{
				UserID:       user,
				ServiceName:  "Netflix",
				PlanID:       plan,
				Price:        299,
				DurationDays: 30,
			})
			s.Require().NoError(err, "покупка должна пройти без ошибки")
			s.Assert().NotZero(int(subID), "должен вернуться ненулевой ID подписки")
		})

		pt.WithNewStep("Проверка: баланс списан до 701", func(s provider.StepCtx) {
			gotUser, err := userRepo.GetUser(context.Background(), user)
			s.Require().NoError(err, "пользователь должен читаться из БД")
			s.Assert().Equal(1000-299, gotUser.Balance, "баланс должен уменьшиться на цену подписки (1000 → 701)")
		})

		pt.WithNewStep("Проверка: записан ровно один платёж-расход", func(s provider.StepCtx) {
			payments, err := paymentRepo.GetUserPayments(context.Background(), user)
			s.Require().NoError(err, "платежи должны читаться из БД")
			s.Assert().Equal(1, len(payments), "должен быть ровно один платёж")
			s.Assert().Equal(299, payments[0].Amount, "сумма платежа должна быть 299")
			s.Assert().Equal(domain.PaymentEXPENCE, payments[0].PaymentType, "тип платежа должен быть expence")
		})
	})
}
