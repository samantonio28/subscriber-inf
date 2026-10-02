//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestSubscriptionPlanRepoIntegration проверяет доступ к данным планов подписок
// против реального PostgreSQL. Каждый подтест независим.
func TestSubscriptionPlanRepoIntegration(t *testing.T) {
	runner.Run(t, "SubscriptionPlanRepo integration", func(pt provider.T) {
		pt.Feature("SubscriptionPlanRepo")
		pt.Tags("integration", "data-access", "plans")

		pool := testdb.Connect(pt)
		repo, err := NewSubscriptionPlanRepo(pool)
		pt.Require().NoError(err)
		ctx := context.Background()

		seedSvc := func(t testing.TB) int {
			t.Helper()
			testdb.Reset(t, pool)
			return testdb.SeedService(t, pool, "YouTube")
		}

		pt.Run("Create and GetByID", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499})
			t.Require().NoError(err)
			t.Assert().NotZero(int(id))

			got, err := repo.GetByID(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal("YouTube Premium", got.Name)
			t.Assert().Equal(499, got.Price)
		})

		pt.Run("GetByService returns plans of service", func(t provider.T) {
			svc := seedSvc(t)
			_, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Family", DurationDays: 30, Price: 899})
			t.Require().NoError(err)
			list, err := repo.GetByService(ctx, svc)
			t.Require().NoError(err)
			t.Assert().Equal(1, len(list))
		})

		pt.Run("GetAll returns all plans", func(t provider.T) {
			svc := seedSvc(t)
			_, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499})
			t.Require().NoError(err)
			list, err := repo.GetAll(ctx)
			t.Require().NoError(err)
			t.Assert().Equal(1, len(list))
		})

		pt.Run("Update modifies plan", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 299})
			t.Require().NoError(err)
			err = repo.Update(ctx, domain.SubscriptionPlan{PlanID: id, ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 349})
			t.Require().NoError(err)
			got, err := repo.GetByID(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal(349, got.Price)
		})

		pt.Run("Delete removes plan", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Music", DurationDays: 30, Price: 199})
			t.Require().NoError(err)
			t.Require().NoError(repo.Delete(ctx, id))
			_, err = repo.GetByID(ctx, id)
			t.Assert().Error(err)
		})
	})
}
