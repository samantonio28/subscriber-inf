package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

// testDBURL возвращает DSN для тестовой базы данных.
// По умолчанию используется localhost:8001, база dev.
const testDBURL = "postgres://postgres:secret@localhost:8001/dev?sslmode=disable"

// setupTestDB создаёт пул подключений к тестовой БД.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDBURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})
	return pool
}

// truncateTables очищает таблицы, используемые в тестах.
func truncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE subscriptions, services, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// firstOfMonth возвращает дату, которая является первым числом месяца.
// Если переданная дата уже первое число, возвращает её, иначе перемещает на первое число того же месяца.
func firstOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

// sameDate проверяет, что две даты (time.Time) соответствуют одному календарному дню (год, месяц, день).
func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// createTestUser создаёт тестового пользователя в БД и возвращает его UUID.
func createTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO users (user_id, email, password, user_name, age, balance)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID,
		fmt.Sprintf("%s@test.com", userID.String()[:8]),
		"password123",
		"Test User",
		25,
		1000,
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return userID
}

func TestSubRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	truncateTables(t, pool)

	repo, err := NewSubRepo(pool)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	if repo == nil {
		t.Fatal("repo is nil")
	}

	ctx := context.Background()

	t.Run("StoreSub and retrieve by ID", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)
		// Сервис будет создан автоматически через PutServiceName в StoreSub

		now := time.Now()
		startDate := firstOfMonth(now)
		endDate := firstOfMonth(now.AddDate(0, 1, 0)) // первое число следующего месяца

		sub := domain.Subscription{
			UserID:      userID,
			ServiceName: "Netflix",
			Price:       299,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		// Store
		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Errorf("StoreSub failed: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero ID")
		}

		// Retrieve
		retrieved, err := repo.Sub(ctx, id)
		if err != nil {
			t.Errorf("Sub failed: %v", err)
		}
		if retrieved.UserID != sub.UserID {
			t.Errorf("UserID mismatch: got %v, want %v", retrieved.UserID, sub.UserID)
		}
		if retrieved.ServiceName != sub.ServiceName {
			t.Errorf("ServiceName mismatch: got %s, want %s", retrieved.ServiceName, sub.ServiceName)
		}
		if retrieved.Price != sub.Price {
			t.Errorf("Price mismatch: got %d, want %d", retrieved.Price, sub.Price)
		}
		if retrieved.SubType != sub.SubType {
			t.Errorf("SubType mismatch: got %v, want %v", retrieved.SubType, sub.SubType)
		}
		if !sameDate(retrieved.StartDate, sub.StartDate) {
			t.Errorf("StartDate mismatch: got %v, want %v", retrieved.StartDate, sub.StartDate)
		}
		if !sameDate(retrieved.EndDate, sub.EndDate) {
			t.Errorf("EndDate mismatch: got %v, want %v", retrieved.EndDate, sub.EndDate)
		}
	})

	t.Run("StoreSub with end date equal to start date", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)

		startDate := firstOfMonth(time.Now())
		endDate := startDate // допустимо, т.к. end_date >= start_date

		sub := domain.Subscription{
			UserID:      userID,
			ServiceName: "Spotify",
			Price:       129,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Errorf("StoreSub failed: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero ID")
		}

		retrieved, err := repo.Sub(ctx, id)
		if err != nil {
			t.Errorf("Sub failed: %v", err)
		}
		if !sameDate(retrieved.EndDate, sub.EndDate) {
			t.Errorf("EndDate mismatch: got %v, want %v", retrieved.EndDate, sub.EndDate)
		}
	})

	t.Run("UserSubs returns subscriptions for user", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)

		startDate := firstOfMonth(time.Now())
		sub1 := domain.Subscription{
			UserID:      userID,
			ServiceName: "Service A",
			Price:       100,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     firstOfMonth(startDate.AddDate(0, 1, 0)),
		}
		sub2 := domain.Subscription{
			UserID:      userID,
			ServiceName: "Service B",
			Price:       200,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     firstOfMonth(startDate.AddDate(0, 2, 0)),
		}

		id1, err := repo.StoreSub(ctx, sub1)
		if err != nil {
			t.Errorf("StoreSub 1 failed: %v", err)
		}
		id2, err := repo.StoreSub(ctx, sub2)
		if err != nil {
			t.Errorf("StoreSub 2 failed: %v", err)
		}

		subs, err := repo.UserSubs(ctx, userID)
		if err != nil {
			t.Errorf("UserSubs failed: %v", err)
		}
		if len(subs) != 2 {
			t.Errorf("expected 2 subscriptions, got %d", len(subs))
		}

		// Проверяем, что оба ID присутствуют
		found1, found2 := false, false
		for _, s := range subs {
			if s.SubId == id1 {
				found1 = true
			}
			if s.SubId == id2 {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Errorf("missing subscriptions: found1=%v, found2=%v", found1, found2)
		}
	})

	t.Run("UpdateSub modifies existing subscription", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)

		startDate := firstOfMonth(time.Now())
		sub := domain.Subscription{
			UserID:      userID,
			ServiceName: "OldService",
			Price:       100,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     firstOfMonth(startDate.AddDate(0, 1, 0)),
		}
		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Errorf("StoreSub failed: %v", err)
		}

		// Обновляем
		updated := domain.Subscription{
			SubId:       id,
			UserID:      userID,
			ServiceName: "NewService",
			Price:       150,
			SubType:     domain.SubTypePromocode,
			StartDate:   startDate,
			EndDate:     firstOfMonth(startDate.AddDate(0, 2, 0)),
		}
		err = repo.UpdateSub(ctx, updated)
		if err != nil {
			t.Errorf("UpdateSub failed: %v", err)
		}

		// Проверяем
		retrieved, err := repo.Sub(ctx, id)
		if err != nil {
			t.Errorf("Sub failed: %v", err)
		}
		if retrieved.ServiceName != updated.ServiceName {
			t.Errorf("ServiceName mismatch: got %s, want %s", retrieved.ServiceName, updated.ServiceName)
		}
		if retrieved.Price != updated.Price {
			t.Errorf("Price mismatch: got %d, want %d", retrieved.Price, updated.Price)
		}
		if retrieved.SubType != updated.SubType {
			t.Errorf("SubType mismatch: got %v, want %v", retrieved.SubType, updated.SubType)
		}
	})

	t.Run("DeleteSub removes subscription", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)

		startDate := firstOfMonth(time.Now())
		sub := domain.Subscription{
			UserID:      userID,
			ServiceName: "ToDelete",
			Price:       50,
			SubType:     domain.SubTypeUsual,
			StartDate:   startDate,
			EndDate:     firstOfMonth(startDate.AddDate(0, 1, 0)),
		}
		id, err := repo.StoreSub(ctx, sub)
		if err != nil {
			t.Errorf("StoreSub failed: %v", err)
		}

		// Удаляем
		err = repo.DeleteSub(ctx, id)
		if err != nil {
			t.Errorf("DeleteSub failed: %v", err)
		}

		// Проверяем, что подписка больше не существует
		_, err = repo.Sub(ctx, id)
		if err == nil {
			t.Error("expected error after deletion, got nil")
		}
	})

	t.Run("Sub returns error for non-existent ID", func(t *testing.T) {
		truncateTables(t, pool)

		_, err := repo.Sub(ctx, domain.SubID(99999))
		if err == nil {
			t.Error("expected error for non-existent ID, got nil")
		}
	})

	t.Run("UpdateSub returns error for non-existent subscription", func(t *testing.T) {
		truncateTables(t, pool)
		userID := createTestUser(t, pool)

		nonExistent := domain.Subscription{
			SubId:       domain.SubID(99999),
			UserID:      userID,
			ServiceName: "Ghost",
			Price:       0,
			SubType:     domain.SubTypeUsual,
			StartDate:   firstOfMonth(time.Now()),
			EndDate:     firstOfMonth(time.Now().AddDate(0, 1, 0)),
		}
		err := repo.UpdateSub(ctx, nonExistent)
		if err == nil {
			t.Error("expected error for non-existent subscription, got nil")
		}
	})
}
