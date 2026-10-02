//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestUserServiceRepoIntegration проверяет связь пользователь ↔ сервис
// (избранное) против реального PostgreSQL. Каждый подтест независим.
func TestUserServiceRepoIntegration(t *testing.T) {
	runner.Run(t, "UserServiceRepo integration", func(pt provider.T) {
		pt.Feature("UserServiceRepo")
		pt.Description("Доступ к данным связи «пользователь ↔ сервис» (избранное) против реального PostgreSQL: " +
			"добавление, проверка наличия, списки в обе стороны и удаление.")

		pool := testdb.Connect(pt)
		repo, err := NewUserServiceRepo(pool)
		pt.Require().NoError(err, "UserServiceRepo должен создаться")
		ctx := context.Background()

		seed := func(t testing.TB) (uuid.UUID, int) {
			t.Helper()
			testdb.Reset(t, pool)
			user := testdb.SeedUser(t, pool, "fav@example.com", 0, "user")
			svc := testdb.SeedService(t, pool, "Netflix")
			return user, svc
		}

		pt.Run("Add and Exists", func(t provider.T) {
			t.Description("Добавление сервиса в избранное и проверка наличия связи.")
			var user uuid.UUID
			var svc int
			t.WithNewStep("Подготовка и добавление в избранное", func(s provider.StepCtx) {
				user, svc = seed(t)
				s.Require().NoError(repo.Add(ctx, user, svc), "Add не должен вернуть ошибку")
			})
			t.WithNewStep("Проверка наличия связи", func(s provider.StepCtx) {
				ok, err := repo.Exists(ctx, user, svc)
				s.Require().NoError(err, "Exists не должен вернуть ошибку")
				s.Assert().True(ok, "связь должна существовать")
			})
		})

		pt.Run("ListByUser returns service ids", func(t provider.T) {
			t.Description("Список избранных сервисов пользователя.")
			var user uuid.UUID
			var svc int
			t.WithNewStep("Подготовка и добавление", func(s provider.StepCtx) {
				user, svc = seed(t)
				s.Require().NoError(repo.Add(ctx, user, svc), "Add не должен вернуть ошибку")
			})
			t.WithNewStep("Получение списка сервисов пользователя", func(s provider.StepCtx) {
				ids, err := repo.ListByUser(ctx, user)
				s.Require().NoError(err, "ListByUser не должен вернуть ошибку")
				s.Assert().Equal([]int{svc}, ids, "должен вернуться список с единственным service_id")
			})
		})

		pt.Run("ListByService returns user ids", func(t provider.T) {
			t.Description("Список пользователей, добавивших сервис в избранное.")
			var user uuid.UUID
			var svc int
			t.WithNewStep("Подготовка и добавление", func(s provider.StepCtx) {
				user, svc = seed(t)
				s.Require().NoError(repo.Add(ctx, user, svc), "Add не должен вернуть ошибку")
			})
			t.WithNewStep("Получение списка пользователей сервиса", func(s provider.StepCtx) {
				users, err := repo.ListByService(ctx, svc)
				s.Require().NoError(err, "ListByService не должен вернуть ошибку")
				s.Assert().Equal([]uuid.UUID{user}, users, "должен вернуться список с единственным user_id")
			})
		})

		pt.Run("Remove deletes relationship", func(t provider.T) {
			t.Description("Удаление сервиса из избранного: связь должна исчезнуть.")
			var user uuid.UUID
			var svc int
			t.WithNewStep("Подготовка, добавление и удаление", func(s provider.StepCtx) {
				user, svc = seed(t)
				s.Require().NoError(repo.Add(ctx, user, svc), "Add не должен вернуть ошибку")
				s.Require().NoError(repo.Remove(ctx, user, svc), "Remove не должен вернуть ошибку")
			})
			t.WithNewStep("Проверка отсутствия связи", func(s provider.StepCtx) {
				ok, err := repo.Exists(ctx, user, svc)
				s.Require().NoError(err, "Exists не должен вернуть ошибку")
				s.Assert().False(ok, "связь должна быть удалена")
			})
		})
	})
}
