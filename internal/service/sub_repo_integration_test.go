//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestSubRepoIntegration проверяет компонент доступа к данным (SubRepo) против
// реального PostgreSQL: полный жизненный цикл подписки и расчёт стоимости.
//
// Каждый подтест сам сбрасывает БД и сидирует нужные сущности, поэтому
// подтесты независимы друг от друга (устойчивы к -shuffle, повторному запуску
// и запуску по отдельности через -run).
func TestSubRepoIntegration(t *testing.T) {
	runner.Run(t, "SubRepo integration", func(pt provider.T) {
		pt.Feature("SubRepo")
		pt.Tags("integration", "data-access", "subscriptions")

		pool := testdb.Connect(pt)
		repo, err := NewSubRepo(pool)
		pt.Require().NoError(err)
		ctx := context.Background()

		// seedSubContext очищает БД и создаёт service → plan → user, возвращает их ID.
		seedSubContext := func(t testing.TB) (int, uuid.UUID) {
			t.Helper()
			testdb.Reset(t, pool)
			svc := testdb.SeedService(t, pool, "Netflix")
			plan := testdb.SeedPlan(t, pool, svc, "Netflix Basic", 30, 299)
			user := testdb.SeedUser(t, pool, "user@example.com", 1000, "user")
			return plan, user
		}

		start := testdb.FirstOfMonth(time.Now())
		end := testdb.FirstOfMonth(start.AddDate(0, 1, 0))

		pt.Run("StoreSub and retrieve by ID", func(t provider.T) {
			plan, user := seedSubContext(t)
			sub := domain.Subscription{
				UserID:      user,
				ServiceName: "Netflix", // имя выводится через join plan → service
				Price:       299,
				SubType:     domain.SubTypeUsual,
				StartDate:   start,
				EndDate:     end,
				PlanID:      plan,
			}

			id, err := repo.StoreSub(ctx, sub)
			t.Require().NoError(err)
			t.Assert().NotZero(int(id))

			got, err := repo.Sub(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal(user, got.UserID)
			t.Assert().Equal(299, got.Price)
			t.Assert().Equal("Netflix", got.ServiceName)
			t.Assert().Equal(plan, got.PlanID)
		})

		pt.Run("UserSubs returns subscriptions for user", func(t provider.T) {
			plan, user := seedSubContext(t)
			sub1 := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
			sub2 := domain.Subscription{UserID: user, Price: 200, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
			_, err := repo.StoreSub(ctx, sub1)
			t.Require().NoError(err)
			_, err = repo.StoreSub(ctx, sub2)
			t.Require().NoError(err)

			subs, err := repo.UserSubs(ctx, user)
			t.Require().NoError(err)
			t.Assert().Equal(2, len(subs))
		})

		pt.Run("UpdateSub modifies existing subscription", func(t provider.T) {
			plan, user := seedSubContext(t)
			sub := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
			id, err := repo.StoreSub(ctx, sub)
			t.Require().NoError(err)

			updated := domain.Subscription{
				SubId:       id,
				UserID:      user,
				Price:       399,
				SubType:     domain.SubTypePromocode,
				StartDate:   start,
				EndDate:     end,
				PlanID:      plan,
			}
			t.Require().NoError(repo.UpdateSub(ctx, updated))

			got, err := repo.Sub(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal(399, got.Price)
			t.Assert().Equal("promocode", got.SubType.String())
		})

		pt.Run("DeleteSub removes subscription", func(t provider.T) {
			plan, user := seedSubContext(t)
			sub := domain.Subscription{UserID: user, Price: 50, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan}
			id, err := repo.StoreSub(ctx, sub)
			t.Require().NoError(err)

			t.Require().NoError(repo.DeleteSub(ctx, id))
			_, err = repo.Sub(ctx, id)
			t.Assert().Error(err)
		})

		pt.Run("Sub returns error for non-existent ID", func(t provider.T) {
			seedSubContext(t)
			_, err := repo.Sub(ctx, domain.SubID(999999))
			t.Assert().Error(err)
		})

		pt.Run("SubsTotalCosts computes cost for period", func(t provider.T) {
			plan, user := seedSubContext(t)
			jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
			mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
			sub := domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: jan, EndDate: feb, PlanID: plan}
			_, err := repo.StoreSub(ctx, sub)
			t.Require().NoError(err)

			sum, subs, err := repo.SubsTotalCosts(ctx, domain.SubsFilter{
				StartDate:   jan,
				EndDate:     mar,
				UserID:      user,
				ServiceName: "Netflix",
				SubType:     domain.SubTypeUsual,
			})
			t.Require().NoError(err)
			t.Assert().Equal(100, sum)
			t.Assert().Equal(1, len(subs))
		})
	})
}
