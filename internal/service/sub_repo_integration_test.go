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
func TestSubRepoIntegration(t *testing.T) {
	runner.Run(t, "SubRepo integration", func(pt provider.T) {
		pt.Feature("SubRepo")
		pt.Description("Доступ к данным подписок (SubRepo) против реального PostgreSQL: " +
			"создание, чтение, список, обновление, удаление и расчёт стоимости за период.")

		pool := testdb.Connect(pt)
		repo, err := NewSubRepo(pool)
		pt.Require().NoError(err, "SubRepo должен создаться")
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
			t.Description("Создание подписки через StoreSub и её чтение через Sub() по ID, с проверкой всех полей.")
			var plan int
			var user uuid.UUID

			t.WithNewStep("Подготовка: service → plan → user", func(s provider.StepCtx) {
				plan, user = seedSubContext(t)
			})

			t.WithNewStep("Создание подписки и её чтение", func(s provider.StepCtx) {
				sub := domain.Subscription{
					UserID:      user,
					ServiceName: "Netflix",
					Price:       299,
					SubType:     domain.SubTypeUsual,
					StartDate:   start,
					EndDate:     end,
					PlanID:      plan,
				}
				id, err := repo.StoreSub(ctx, sub)
				s.Require().NoError(err, "StoreSub не должен вернуть ошибку")
				s.Assert().NotZero(int(id), "ID подписки должен быть ненулевым")

				got, err := repo.Sub(ctx, id)
				s.Require().NoError(err, "Sub() не должен вернуть ошибку")
				s.Assert().Equal(user, got.UserID, "user_id должен совпадать с созданным")
				s.Assert().Equal(299, got.Price, "цена должна быть 299")
				s.Assert().Equal("Netflix", got.ServiceName, "имя сервиса должно выводиться из plan")
				s.Assert().Equal(plan, got.PlanID, "plan_id должен совпадать")
			})
		})

		pt.Run("UserSubs returns subscriptions for user", func(t provider.T) {
			t.Description("Список подписок пользователя через UserSubs: должны вернуться обе созданные подписки.")
			var plan int
			var user uuid.UUID

			t.WithNewStep("Подготовка: user и две подписки", func(s provider.StepCtx) {
				plan, user = seedSubContext(t)
				_, err := repo.StoreSub(ctx, domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan})
				s.Require().NoError(err, "первая подписка должна создаться")
				_, err = repo.StoreSub(ctx, domain.Subscription{UserID: user, Price: 200, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan})
				s.Require().NoError(err, "вторая подписка должна создаться")
			})

			t.WithNewStep("Получение списка подписок", func(s provider.StepCtx) {
				subs, err := repo.UserSubs(ctx, user)
				s.Require().NoError(err, "UserSubs не должен вернуть ошибку")
				s.Assert().Equal(2, len(subs), "должно вернуться ровно 2 подписки")
			})
		})

		pt.Run("UpdateSub modifies existing subscription", func(t provider.T) {
			t.Description("Обновление существующей подписки: цена и тип меняются, user_id остаётся прежним.")
			var plan int
			var user uuid.UUID
			var id domain.SubID

			t.WithNewStep("Подготовка: подписка ценой 100", func(s provider.StepCtx) {
				plan, user = seedSubContext(t)
				var err error
				id, err = repo.StoreSub(ctx, domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan})
				s.Require().NoError(err, "подписка должна создаться")
			})

			t.WithNewStep("Обновление подписки (цена 399, тип promocode)", func(s provider.StepCtx) {
				err := repo.UpdateSub(ctx, domain.Subscription{
					SubId: id, UserID: user, Price: 399, SubType: domain.SubTypePromocode,
					StartDate: start, EndDate: end, PlanID: plan,
				})
				s.Require().NoError(err, "UpdateSub не должен вернуть ошибку")

				got, err := repo.Sub(ctx, id)
				s.Require().NoError(err, "подписка должна читаться после обновления")
				s.Assert().Equal(399, got.Price, "цена должна стать 399")
				s.Assert().Equal("promocode", got.SubType.String(), "тип должен стать promocode")
			})
		})

		pt.Run("DeleteSub removes subscription", func(t provider.T) {
			t.Description("Удаление подписки: после DeleteSub чтение по ID должно вернуть ошибку.")
			var plan int
			var user uuid.UUID
			var id domain.SubID

			t.WithNewStep("Подготовка: подписка", func(s provider.StepCtx) {
				plan, user = seedSubContext(t)
				var err error
				id, err = repo.StoreSub(ctx, domain.Subscription{UserID: user, Price: 50, SubType: domain.SubTypeUsual, StartDate: start, EndDate: end, PlanID: plan})
				s.Require().NoError(err, "подписка должна создаться")
			})

			t.WithNewStep("Удаление и проверка отсутствия", func(s provider.StepCtx) {
				s.Require().NoError(repo.DeleteSub(ctx, id), "DeleteSub не должен вернуть ошибку")
				_, err := repo.Sub(ctx, id)
				s.Assert().Error(err, "чтение удалённой подписки должно вернуть ошибку")
			})
		})

		pt.Run("Sub returns error for non-existent ID", func(t provider.T) {
			t.Description("Чтение подписки по несуществующему ID должно вернуть ошибку.")
			t.WithNewStep("Подготовка: пустая БД", func(s provider.StepCtx) {
				seedSubContext(t)
			})
			t.WithNewStep("Чтение по ID 999999", func(s provider.StepCtx) {
				_, err := repo.Sub(ctx, domain.SubID(999999))
				s.Assert().Error(err, "должна вернуться ошибка для несуществующего ID")
			})
		})

		pt.Run("SubsTotalCosts computes cost for period", func(t provider.T) {
			t.Description("Расчёт стоимости за период: подписка 100 в месяц за январь должна дать сумму 100.")
			var plan int
			var user uuid.UUID
			jan := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			feb := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
			mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

			t.WithNewStep("Подготовка: подписка 100 на январь", func(s provider.StepCtx) {
				plan, user = seedSubContext(t)
				_, err := repo.StoreSub(ctx, domain.Subscription{UserID: user, Price: 100, SubType: domain.SubTypeUsual, StartDate: jan, EndDate: feb, PlanID: plan})
				s.Require().NoError(err, "подписка должна создаться")
			})

			t.WithNewStep("Расчёт стоимости за январь-март", func(s provider.StepCtx) {
				sum, subs, err := repo.SubsTotalCosts(ctx, domain.SubsFilter{
					StartDate: jan, EndDate: mar, UserID: user, ServiceName: "Netflix", SubType: domain.SubTypeUsual,
				})
				s.Require().NoError(err, "SubsTotalCosts не должен вернуть ошибку")
				s.Assert().Equal(100, sum, "сумма за месяц должна быть 100")
				s.Assert().Equal(1, len(subs), "в результат должна попасть одна подписка")
			})
		})
	})
}
