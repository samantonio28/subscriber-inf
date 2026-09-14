package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgPool — минимальный интерфейс пула соединений PostgreSQL, который
// используют репозитории. *pgxpool.Pool удовлетворяет этому интерфейсу.
//
// Зачем нужен: unit-тесты в "классическом" стиле должны иметь возможность
// подменить PostgreSQL, не поднимая реальную БД. Через этот интерфейс в
// настоящие структуры репозиториев передаётся «замоканный» коннект
// (см. internal/testutil/pgxstubs.go и classic-тесты в internal/usecase).
type PgPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}