//go:build integration

package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestApplyPromocodeIntegration проверяет бизнес-логику применения промокода к
// активной подписке на реальной БД: пересчёт цены, обновление подписки и
// инкремент счётчика использований.
func TestApplyPromocodeIntegration(t *testing.T) {
	runner.Run(t, "ApplyPromocode integration", func(pt provider.T) {
		pt.Feature("ApplyPromocode")
		pt.Description("Бизнес-логика применения промокода к активной подписке на реальной БД: " +
			"цена пересчитывается со скидкой, подписка обновляется, счётчик использований промокода инкрементируется.")

		var (
			subRepo   domain.SubscriptionRepository
			promoRepo domain.PromocodeRepository
			uc        *ApplyPromocodeUC
			subID     int
			promoID   int
		)

		pool := testdb.Connect(pt)
		testdb.Reset(pt, pool)

		pt.WithNewStep("Сборка репозиториев и usecase", func(s provider.StepCtx) {
			var err error
			subRepo, err = service.NewSubRepo(pool)
			s.Require().NoError(err, "SubRepo должен создаться")
			promoRepo, err = service.NewPromocodeRepo(pool)
			s.Require().NoError(err, "PromocodeRepo должен создаться")
			uc, err = NewApplyPromocodeUC(subRepo, promoRepo, &testutil.NopLogger{})
			s.Require().NoError(err, "ApplyPromocodeUC должен создаться")
		})

		pt.WithNewStep("Подготовка данных: активная подписка ценой 1000 и промокод SAVE20", func(s provider.StepCtx) {
			svc := testdb.SeedService(pt, pool, "Netflix")
			plan := testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 1000)
			user := testdb.SeedUser(pt, pool, "subscriber@example.com", 0, "user")
			start := testdb.FirstOfMonth(time.Now())
			end := testdb.FirstOfMonth(start.AddDate(0, 1, 0))
			subID = testdb.SeedSubscription(pt, pool, user, plan, 1000, "usual", start, end)
			promoID = testdb.SeedPromocode(pt, pool, svc, "SAVE20", 20, 5)
			s.Assert().NotZero(subID, "подписка должна создаться")
			s.Assert().NotZero(promoID, "промокод должен создаться")
		})

		pt.WithNewStep("Применение промокода SAVE20", func(s provider.StepCtx) {
			out, err := uc.Apply(context.Background(), ApplyPromocodeInput{
				SubscriptionID: subID,
				PromocodeValue: "SAVE20",
			})
			s.Require().NoError(err, "применение промокода должно пройти без ошибки")
			s.Assert().Equal(20, out.DiscountApplied, "скидка должна быть 20%")
			s.Assert().Equal(800, out.NewPrice, "новая цена должна быть 1000 * 80% = 800")
		})

		pt.WithNewStep("Проверка: подписка обновлена (цена 800, тип promocode)", func(s provider.StepCtx) {
			sub, err := subRepo.Sub(context.Background(), domain.SubID(subID))
			s.Require().NoError(err, "подписка должна читаться из БД")
			s.Assert().Equal(800, sub.Price, "цена подписки должна стать 800")
			s.Assert().Equal("promocode", sub.SubType.String(), "тип подписки должен стать promocode")
		})

		pt.WithNewStep("Проверка: счётчик использований промокода = 1", func(s provider.StepCtx) {
			promo, err := promoRepo.GetByID(context.Background(), domain.PromocodeID(promoID))
			s.Require().NoError(err, "промокод должен читаться из БД")
			s.Assert().Equal(1, promo.CurUses, "счётчик использований должен увеличиться до 1")
		})
	})
}
