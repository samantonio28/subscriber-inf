// Package testdb содержит хелперы для интеграционных и E2E-тестов, которые
// работают с реальным инстансом PostgreSQL (тестовый стенд из
// docker-compose.test.yml), а не с заглушками.
//
// Зачем отдельный пакет: unit-тесты используют «замоканный» PostgreSQL
// (internal/testutil/pgxstubs.go), а интеграционные/E2E тесты обязаны
// взаимодействовать с настоящим хранилищем (требование ЛР №2). Здесь собраны:
//   - подключение к тестовой БД (с возможностью переопределить DSN),
//   - сброс состояния БД между прогонами (идемпотентность),
//   - сидирование базовых сущностей (service → plan → subscription → promocode).
package testdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultDSN — адрес тестового инстанса PostgreSQL по умолчанию.
// Соответствует сервису postgres_test из docker-compose.test.yml.
const DefaultDSN = "postgres://postgres:secret@localhost:8002/test?sslmode=disable"

// DSN возвращает строку подключения к тестовой БД.
// Значение можно переопределить переменной окружения TEST_DATABASE_URL —
// это позволяет нескольким разработчикам запускать тесты одновременно,
// каждый на своём инстансе, не мешая друг другу (требование ЛР №2).
func DSN() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return DefaultDSN
}

// Connect открывает пул подключений к тестовой БД и регистрирует его закрытие
// в t.Cleanup.
func Connect(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), DSN())
	if err != nil {
		t.Fatalf("connect to test db %q: %v", DSN(), err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// Reset очищает все таблицы тестовой БД и сбрасывает identity-счётчики.
// Благодаря этому каждый тест стартует с чистого состояния и его можно
// запускать многократно без накопления данных (требования: идемпотентность,
// откат хранилища, целостность при прерывании).
func Reset(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	if err := TruncateAll(context.Background(), pool); err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// Setup одной строкой «поднимает» соединение с тестовой БД и сбрасывает её к
// чистому состоянию: подключается к инстансу хранилища и очищает все таблицы.
// Возвращает пул подключений, готовый к работе теста.
func Setup(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool := Connect(t)
	Reset(t, pool)
	return pool
}

// TruncateAll очищает все таблицы (TRUNCATE ... CASCADE RESTART IDENTITY).
// Порядок не важен: CASCADE обрывает внешние ключи, а полный список таблиц
// гарантирует, что ничего не останется.
func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
		TRUNCATE TABLE
			user_referrals,
			payments,
			cards,
			subscriptions,
			promocodes,
			subscription_plans,
			user_services,
			services,
			users
		RESTART IDENTITY CASCADE`
	_, err := pool.Exec(ctx, q)
	return err
}

// SeedService создаёт сервис и возвращает его service_id.
func SeedService(t testing.TB, pool *pgxpool.Pool, name string) int {
	t.Helper()
	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO services (service_name) VALUES ($1) RETURNING service_id`,
		name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed service %q: %v", name, err)
	}
	return id
}

// SeedPlan создаёт план подписки для сервиса и возвращает plan_id.
func SeedPlan(t testing.TB, pool *pgxpool.Pool, serviceID int, name string, durationDays, price int) int {
	t.Helper()
	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO subscription_plans (service_id, name, duration_days, price)
		 VALUES ($1, $2, $3, $4) RETURNING plan_id`,
		serviceID, name, durationDays, price,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed plan %q: %v", name, err)
	}
	return id
}

// SeedUser создаёт пользователя и возвращает его user_id.
func SeedUser(t testing.TB, pool *pgxpool.Pool, email string, balance int, role string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if role == "" {
		role = "user"
	}
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (user_id, email, password, user_name, age, balance, referral_code, role)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, email, "secret", "Test User", 25, balance, "REF"+id.String()[:6], role,
	)
	if err != nil {
		t.Fatalf("seed user %q: %v", email, err)
	}
	return id
}

// SeedSubscription создаёт подписку (требует существующих user и plan)
// и возвращает sub_id.
func SeedSubscription(t testing.TB, pool *pgxpool.Pool, userID uuid.UUID, planID int, price int, subType string, startDate, endDate time.Time) int {
	t.Helper()
	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO subscriptions (user_id, plan_id, price, sub_type, start_date, end_date)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING sub_id`,
		userID, planID, price, subType, startDate, endDate,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed subscription: %v", err)
	}
	return id
}

// SeedPromocode создаёт промокод для сервиса и возвращает promocode_id.
func SeedPromocode(t testing.TB, pool *pgxpool.Pool, serviceID int, code string, discount, maxUses int) int {
	t.Helper()
	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO promocodes (service_id, promocode, discount, max_uses, cur_uses, status, duration_days, expires_at)
		 VALUES ($1, $2, $3, $4, 0, 'ACTIVE', 30, now() + interval '30 days') RETURNING promocode_id`,
		serviceID, code, discount, maxUses,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed promocode %q: %v", code, err)
	}
	return id
}

// FirstOfMonth возвращает первое число текущего месяца (полезно для полей
// start_date/end_date, на которые наложен CHECK «день = 1»).
func FirstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
