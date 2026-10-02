//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

func newPromocode(svc int, value string, discount, maxUses int) domain.Promocode {
	return domain.Promocode{
		ServiceID:    svc,
		Value:        value,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		Discount:     discount,
		MaxUses:      maxUses,
		CurUses:      0,
		Status:       domain.PromocodeStatusActive,
		DurationDays: 30,
	}
}

// TestPromocodeRepoIntegration проверяет доступ к данным промокодов против
// реального PostgreSQL. Каждый подтест независим (свой сброс и сидирование).
func TestPromocodeRepoIntegration(t *testing.T) {
	runner.Run(t, "PromocodeRepo integration", func(pt provider.T) {
		pt.Feature("PromocodeRepo")
		pt.Tags("integration", "data-access", "promocodes")

		pool := testdb.Connect(pt)
		repo, err := NewPromocodeRepo(pool)
		pt.Require().NoError(err)
		ctx := context.Background()

		seedSvc := func(t testing.TB) int {
			t.Helper()
			testdb.Reset(t, pool)
			return testdb.SeedService(t, pool, "Spotify")
		}

		pt.Run("Create and GetByID", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, newPromocode(svc, "SAVE20", 20, 5))
			t.Require().NoError(err)
			t.Assert().NotZero(int(id))

			got, err := repo.GetByID(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal("SAVE20", got.Value)
			t.Assert().Equal(20, got.Discount)
		})

		pt.Run("GetByCode", func(t provider.T) {
			svc := seedSvc(t)
			_, err := repo.Create(ctx, newPromocode(svc, "WELCOME5", 5, 10))
			t.Require().NoError(err)
			got, err := repo.GetByCode(ctx, "WELCOME5")
			t.Require().NoError(err)
			t.Assert().Equal("WELCOME5", got.Value)
		})

		pt.Run("GetByService returns promocodes of service", func(t provider.T) {
			svc := seedSvc(t)
			_, err := repo.Create(ctx, newPromocode(svc, "SAVE10", 10, 3))
			t.Require().NoError(err)
			_, err = repo.Create(ctx, newPromocode(svc, "SAVE30", 30, 3))
			t.Require().NoError(err)
			list, err := repo.GetByService(ctx, svc)
			t.Require().NoError(err)
			t.Assert().Equal(2, len(list))
		})

		pt.Run("Update modifies promocode", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, newPromocode(svc, "SAVE40", 40, 4))
			t.Require().NoError(err)
			pc := newPromocode(svc, "SAVE40", 50, 4)
			pc.PromocodeID = id
			t.Require().NoError(repo.Update(ctx, pc))
			got, err := repo.GetByID(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal(50, got.Discount)
		})

		pt.Run("IncrementUses increases cur_uses", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, newPromocode(svc, "SAVE60", 60, 5))
			t.Require().NoError(err)
			t.Require().NoError(repo.IncrementUses(ctx, id))
			got, err := repo.GetByID(ctx, id)
			t.Require().NoError(err)
			t.Assert().Equal(1, got.CurUses)
		})

		pt.Run("Delete removes promocode", func(t provider.T) {
			svc := seedSvc(t)
			id, err := repo.Create(ctx, newPromocode(svc, "SAVE70", 70, 1))
			t.Require().NoError(err)
			t.Require().NoError(repo.Delete(ctx, id))
			_, err = repo.GetByID(ctx, id)
			t.Assert().Error(err)
		})
	})
}
