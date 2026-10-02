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
		pt.Description("Доступ к данным планов подписок (SubscriptionPlanRepo) против реального PostgreSQL: " +
			"создание, чтение по ID и сервису, список, обновление и удаление.")

		pool := testdb.Connect(pt)
		repo, err := NewSubscriptionPlanRepo(pool)
		pt.Require().NoError(err, "SubscriptionPlanRepo должен создаться")
		ctx := context.Background()

		seedSvc := func(t testing.TB) int {
			t.Helper()
			testdb.Reset(t, pool)
			return testdb.SeedService(t, pool, "YouTube")
		}

		pt.Run("Create and GetByID", func(t provider.T) {
			t.Description("Создание плана подписки и его чтение по ID.")
			var svc int
			t.WithNewStep("Подготовка: сервис YouTube", func(s provider.StepCtx) {
				svc = seedSvc(t)
			})
			t.WithNewStep("Создание плана и чтение по ID", func(s provider.StepCtx) {
				id, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499})
				s.Require().NoError(err, "Create не должен вернуть ошибку")
				s.Assert().NotZero(int(id), "ID плана должен быть ненулевым")

				got, err := repo.GetByID(ctx, id)
				s.Require().NoError(err, "GetByID не должен вернуть ошибку")
				s.Assert().Equal("YouTube Premium", got.Name, "имя плана должно быть YouTube Premium")
				s.Assert().Equal(499, got.Price, "цена должна быть 499")
			})
		})

		pt.Run("GetByService returns plans of service", func(t provider.T) {
			t.Description("Список планов сервиса: должен вернуться созданный план.")
			var svc int
			t.WithNewStep("Подготовка: план YouTube Family", func(s provider.StepCtx) {
				svc = seedSvc(t)
				_, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Family", DurationDays: 30, Price: 899})
				s.Require().NoError(err, "план должен создаться")
			})
			t.WithNewStep("Получение планов сервиса", func(s provider.StepCtx) {
				list, err := repo.GetByService(ctx, svc)
				s.Require().NoError(err, "GetByService не должен вернуть ошибку")
				s.Assert().Equal(1, len(list), "должен вернуться 1 план")
			})
		})

		pt.Run("GetAll returns all plans", func(t provider.T) {
			t.Description("Полный список планов: должен вернуться созданный план.")
			var svc int
			t.WithNewStep("Подготовка: один план", func(s provider.StepCtx) {
				svc = seedSvc(t)
				_, err := repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Premium", DurationDays: 30, Price: 499})
				s.Require().NoError(err, "план должен создаться")
			})
			t.WithNewStep("Получение всех планов", func(s provider.StepCtx) {
				list, err := repo.GetAll(ctx)
				s.Require().NoError(err, "GetAll не должен вернуть ошибку")
				s.Assert().Equal(1, len(list), "должен вернуться 1 план")
			})
		})

		pt.Run("Update modifies plan", func(t provider.T) {
			t.Description("Обновление плана: цена меняется с 299 на 349.")
			var svc int
			var id domain.PlanID
			t.WithNewStep("Подготовка: план YouTube Lite за 299", func(s provider.StepCtx) {
				svc = seedSvc(t)
				var err error
				id, err = repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 299})
				s.Require().NoError(err, "план должен создаться")
			})
			t.WithNewStep("Обновление цены до 349", func(s provider.StepCtx) {
				err := repo.Update(ctx, domain.SubscriptionPlan{PlanID: id, ServiceID: svc, Name: "YouTube Lite", DurationDays: 30, Price: 349})
				s.Require().NoError(err, "Update не должен вернуть ошибку")
				got, err := repo.GetByID(ctx, id)
				s.Require().NoError(err, "план должен читаться после обновления")
				s.Assert().Equal(349, got.Price, "цена должна стать 349")
			})
		})

		pt.Run("Delete removes plan", func(t provider.T) {
			t.Description("Удаление плана: после Delete чтение по ID должно вернуть ошибку.")
			var svc int
			var id domain.PlanID
			t.WithNewStep("Подготовка: план YouTube Music", func(s provider.StepCtx) {
				svc = seedSvc(t)
				var err error
				id, err = repo.Create(ctx, domain.SubscriptionPlan{ServiceID: svc, Name: "YouTube Music", DurationDays: 30, Price: 199})
				s.Require().NoError(err, "план должен создаться")
			})
			t.WithNewStep("Удаление и проверка отсутствия", func(s provider.StepCtx) {
				s.Require().NoError(repo.Delete(ctx, id), "Delete не должен вернуть ошибку")
				_, err := repo.GetByID(ctx, id)
				s.Assert().Error(err, "чтение удалённого плана должно вернуть ошибку")
			})
		})
	})
}
