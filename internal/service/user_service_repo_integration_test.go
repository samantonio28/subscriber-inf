//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// TestUserServiceRepoIntegration проверяет связь пользователь ↔ сервис
// (избранное) против реального PostgreSQL. Каждый подтест независим.
func TestUserServiceRepoIntegration(t *testing.T) {
	pool := testdb.Connect(t)
	repo, err := NewUserServiceRepo(pool)
	if err != nil {
		t.Fatalf("NewUserServiceRepo: %v", err)
	}
	ctx := context.Background()

	seed := func(t *testing.T) (uuid.UUID, int) {
		t.Helper()
		testdb.Reset(t, pool)
		user := testdb.SeedUser(t, pool, "fav@example.com", 0, "user")
		svc := testdb.SeedService(t, pool, "Netflix")
		return user, svc
	}

	t.Run("Add and Exists", func(t *testing.T) {
		user, svc := seed(t)
		if err := repo.Add(ctx, user, svc); err != nil {
			t.Fatalf("Add: %v", err)
		}
		ok, err := repo.Exists(ctx, user, svc)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if !ok {
			t.Fatal("expected relationship to exist")
		}
	})

	t.Run("ListByUser returns service ids", func(t *testing.T) {
		user, svc := seed(t)
		if err := repo.Add(ctx, user, svc); err != nil {
			t.Fatalf("Add: %v", err)
		}
		ids, err := repo.ListByUser(ctx, user)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(ids) != 1 || ids[0] != svc {
			t.Errorf("expected [%d], got %v", svc, ids)
		}
	})

	t.Run("ListByService returns user ids", func(t *testing.T) {
		user, svc := seed(t)
		if err := repo.Add(ctx, user, svc); err != nil {
			t.Fatalf("Add: %v", err)
		}
		users, err := repo.ListByService(ctx, svc)
		if err != nil {
			t.Fatalf("ListByService: %v", err)
		}
		if len(users) != 1 || users[0] != user {
			t.Errorf("expected [%v], got %v", user, users)
		}
	})

	t.Run("Remove deletes relationship", func(t *testing.T) {
		user, svc := seed(t)
		if err := repo.Add(ctx, user, svc); err != nil {
			t.Fatalf("Add: %v", err)
		}
		if err := repo.Remove(ctx, user, svc); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		ok, err := repo.Exists(ctx, user, svc)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if ok {
			t.Fatal("expected relationship to be removed")
		}
	})
}
