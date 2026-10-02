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
		pt.Description("Доступ к данным промокодов (PromocodeRepo) против реального PostgreSQL: " +
			"CRUD, поиск по коду и сервису, обновление, инкремент счётчика использований.")

		pool := testdb.Connect(pt)
		repo, err := NewPromocodeRepo(pool)
		pt.Require().NoError(err, "PromocodeRepo должен создаться")
		ctx := context.Background()

		seedSvc := func(t testing.TB) int {
			t.Helper()
			testdb.Reset(t, pool)
			return testdb.SeedService(t, pool, "Spotify")
		}

		pt.Run("Create and GetByID", func(t provider.T) {
			t.Description("Создание промокода и его чтение по ID.")
			var svc int
			t.WithNewStep("Подготовка: сервис Spotify", func(s provider.StepCtx) {
				svc = seedSvc(t)
			})
			t.WithNewStep("Создание промокода SAVE20 и чтение по ID", func(s provider.StepCtx) {
				id, err := repo.Create(ctx, newPromocode(svc, "SAVE20", 20, 5))
				s.Require().NoError(err, "Create не должен вернуть ошибку")
				s.Assert().NotZero(int(id), "ID промокода должен быть ненулевым")

				got, err := repo.GetByID(ctx, id)
				s.Require().NoError(err, "GetByID не должен вернуть ошибку")
				s.Assert().Equal("SAVE20", got.Value, "значение промокода должно быть SAVE20")
				s.Assert().Equal(20, got.Discount, "скидка должна быть 20")
			})
		})

		pt.Run("GetByCode", func(t provider.T) {
			t.Description("Поиск промокода по его коду.")
			var svc int
			t.WithNewStep("Подготовка + поиск по коду WELCOME5", func(s provider.StepCtx) {
				svc = seedSvc(t)
				_, err := repo.Create(ctx, newPromocode(svc, "WELCOME5", 5, 10))
				s.Require().NoError(err, "промокод должен создаться")

				got, err := repo.GetByCode(ctx, "WELCOME5")
				s.Require().NoError(err, "GetByCode не должен вернуть ошибку")
				s.Assert().Equal("WELCOME5", got.Value, "найден должен быть WELCOME5")
			})
		})

		pt.Run("GetByService returns promocodes of service", func(t provider.T) {
			t.Description("Список промокодов сервиса: должны вернуться оба созданных.")
			var svc int
			t.WithNewStep("Подготовка: два промокода сервиса", func(s provider.StepCtx) {
				svc = seedSvc(t)
				_, err := repo.Create(ctx, newPromocode(svc, "SAVE10", 10, 3))
				s.Require().NoError(err, "первый промокод должен создаться")
				_, err = repo.Create(ctx, newPromocode(svc, "SAVE30", 30, 3))
				s.Require().NoError(err, "второй промокод должен создаться")
			})
			t.WithNewStep("Получение промокодов сервиса", func(s provider.StepCtx) {
				list, err := repo.GetByService(ctx, svc)
				s.Require().NoError(err, "GetByService не должен вернуть ошибку")
				s.Assert().Equal(2, len(list), "должно вернуться 2 промокода")
			})
		})

		pt.Run("Update modifies promocode", func(t provider.T) {
			t.Description("Обновление промокода: скидка меняется с 40 на 50.")
			var svc int
			var id domain.PromocodeID
			t.WithNewStep("Подготовка: промокод SAVE40", func(s provider.StepCtx) {
				svc = seedSvc(t)
				var err error
				id, err = repo.Create(ctx, newPromocode(svc, "SAVE40", 40, 4))
				s.Require().NoError(err, "промокод должен создаться")
			})
			t.WithNewStep("Обновление скидки до 50", func(s provider.StepCtx) {
				pc := newPromocode(svc, "SAVE40", 50, 4)
				pc.PromocodeID = id
				s.Require().NoError(repo.Update(ctx, pc), "Update не должен вернуть ошибку")

				got, err := repo.GetByID(ctx, id)
				s.Require().NoError(err, "промокод должен читаться после обновления")
				s.Assert().Equal(50, got.Discount, "скидка должна стать 50")
			})
		})

		pt.Run("IncrementUses increases cur_uses", func(t provider.T) {
			t.Description("Инкремент счётчика использований промокода с 0 до 1.")
			var svc int
			var id domain.PromocodeID
			t.WithNewStep("Подготовка: промокод SAVE60", func(s provider.StepCtx) {
				svc = seedSvc(t)
				var err error
				id, err = repo.Create(ctx, newPromocode(svc, "SAVE60", 60, 5))
				s.Require().NoError(err, "промокод должен создаться")
			})
			t.WithNewStep("Инкремент и проверка счётчика", func(s provider.StepCtx) {
				s.Require().NoError(repo.IncrementUses(ctx, id), "IncrementUses не должен вернуть ошибку")
				got, err := repo.GetByID(ctx, id)
				s.Require().NoError(err, "промокод должен читаться")
				s.Assert().Equal(1, got.CurUses, "счётчик должен стать 1")
			})
		})

		pt.Run("Delete removes promocode", func(t provider.T) {
			t.Description("Удаление промокода: после Delete чтение по ID должно вернуть ошибку.")
			var svc int
			var id domain.PromocodeID
			t.WithNewStep("Подготовка: промокод SAVE70", func(s provider.StepCtx) {
				svc = seedSvc(t)
				var err error
				id, err = repo.Create(ctx, newPromocode(svc, "SAVE70", 70, 1))
				s.Require().NoError(err, "промокод должен создаться")
			})
			t.WithNewStep("Удаление и проверка отсутствия", func(s provider.StepCtx) {
				s.Require().NoError(repo.Delete(ctx, id), "Delete не должен вернуть ошибку")
				_, err := repo.GetByID(ctx, id)
				s.Assert().Error(err, "чтение удалённого промокода должно вернуть ошибку")
			})
		})
	})
}
