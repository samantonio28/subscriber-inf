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
		pt.Tags("integration", "data-access", "favorites")

		pool := testdb.Connect(pt)
		repo, err := NewUserServiceRepo(pool)
		pt.Require().NoError(err)
		ctx := context.Background()

		seed := func(t testing.TB) (uuid.UUID, int) {
			t.Helper()
			testdb.Reset(t, pool)
			user := testdb.SeedUser(t, pool, "fav@example.com", 0, "user")
			svc := testdb.SeedService(t, pool, "Netflix")
			return user, svc
		}

		pt.Run("Add and Exists", func(t provider.T) {
			user, svc := seed(t)
			t.Require().NoError(repo.Add(ctx, user, svc))
			ok, err := repo.Exists(ctx, user, svc)
			t.Require().NoError(err)
			t.Assert().True(ok)
		})

		pt.Run("ListByUser returns service ids", func(t provider.T) {
			user, svc := seed(t)
			t.Require().NoError(repo.Add(ctx, user, svc))
			ids, err := repo.ListByUser(ctx, user)
			t.Require().NoError(err)
			t.Assert().Equal([]int{svc}, ids)
		})

		pt.Run("ListByService returns user ids", func(t provider.T) {
			user, svc := seed(t)
			t.Require().NoError(repo.Add(ctx, user, svc))
			users, err := repo.ListByService(ctx, svc)
			t.Require().NoError(err)
			t.Assert().Equal([]uuid.UUID{user}, users)
		})

		pt.Run("Remove deletes relationship", func(t provider.T) {
			user, svc := seed(t)
			t.Require().NoError(repo.Add(ctx, user, svc))
			t.Require().NoError(repo.Remove(ctx, user, svc))
			ok, err := repo.Exists(ctx, user, svc)
			t.Require().NoError(err)
			t.Assert().False(ok)
		})
	})
}
