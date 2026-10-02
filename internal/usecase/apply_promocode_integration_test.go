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
		pt.Tags("integration", "business-logic", "promocodes")

		pool := testdb.Connect(pt)
		testdb.Reset(pt, pool)

		subRepo, err := service.NewSubRepo(pool)
		pt.Require().NoError(err)
		promoRepo, err := service.NewPromocodeRepo(pool)
		pt.Require().NoError(err)

		uc, err := NewApplyPromocodeUC(subRepo, promoRepo, &testutil.NopLogger{})
		pt.Require().NoError(err)

		svc := testdb.SeedService(pt, pool, "Netflix")
		plan := testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 1000)
		user := testdb.SeedUser(pt, pool, "subscriber@example.com", 0, "user")

		start := testdb.FirstOfMonth(time.Now())
		end := testdb.FirstOfMonth(start.AddDate(0, 1, 0))
		subID := testdb.SeedSubscription(pt, pool, user, plan, 1000, "usual", start, end)

		promoID := testdb.SeedPromocode(pt, pool, svc, "SAVE20", 20, 5)

		out, err := uc.Apply(context.Background(), ApplyPromocodeInput{
			SubscriptionID: subID,
			PromocodeValue: "SAVE20",
		})
		pt.Require().NoError(err)
		pt.Assert().Equal(20, out.DiscountApplied)
		pt.Assert().Equal(800, out.NewPrice)

		// Подписка должна быть обновлена: цена снижена, тип — promocode.
		sub, err := subRepo.Sub(context.Background(), domain.SubID(subID))
		pt.Require().NoError(err)
		pt.Assert().Equal(800, sub.Price)
		pt.Assert().Equal("promocode", sub.SubType.String())

		// Счётчик использований промокода должен увеличиться.
		promo, err := promoRepo.GetByID(context.Background(), domain.PromocodeID(promoID))
		pt.Require().NoError(err)
		pt.Assert().Equal(1, promo.CurUses)
	})
}
