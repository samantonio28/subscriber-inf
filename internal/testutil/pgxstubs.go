// Package testutil — вспомогательные инструменты для unit-тестов.
package testutil

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// ---------------------------------------------------------------------------
// Дублёры уровня «замоканного PostgreSQL».
//
// В классических тестах мы НЕ поднимаем реальную БД и НЕ используем
// gomock-моки доменных репозиториев. Вместо этого настоящие структуры
// репозиториев (service.NewSubRepo и т.д.) получают FakePgPool — заглушку
// коннекта к PostgreSQL, а сами строки/транзакции имитируются типами
// StubRow / StubRows / StubTx. Репозитории «думают», что работают с
// настоящим сервером, хотя никакой сети и БД нет.
// ---------------------------------------------------------------------------

// StubRow имитирует pgx.Row — одну строку результата запроса.
type StubRow struct {
	values []any
	err    error
}

// NewStubRow создаёт строку с заданными значениями колонок.
func NewStubRow(values ...any) *StubRow {
	return &StubRow{values: values}
}

// WithError превращает строку в «ошибочную» (Scan вернёт ошибку).
func (r *StubRow) WithError(err error) *StubRow {
	r.err = err
	return r
}

// Scan присваивает значения колонок в переданные указатели.
func (r *StubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("stub row: expected %d destinations, got %d values", len(dest), len(r.values))
	}
	for i := range dest {
		if err := assignValue(r.values[i], dest[i]); err != nil {
			return fmt.Errorf("stub row scan column %d: %w", i, err)
		}
	}
	return nil
}

// StubRows имитирует pgx.Rows — набор строк результата запроса.
type StubRows struct {
	data [][]any
	idx  int
	err  error
}

// NewStubRows создаёт набор строк; каждая строка — список значений колонок.
func NewStubRows(data ...[]any) *StubRows {
	return &StubRows{data: data, idx: -1}
}

// WithError задаёт ошибку, которую вернёт Err().
func (r *StubRows) WithError(err error) *StubRows {
	r.err = err
	return r
}

func (r *StubRows) Next() bool {
	r.idx++
	return r.idx < len(r.data)
}

func (r *StubRows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.idx < 0 || r.idx >= len(r.data) {
		return pgx.ErrNoRows
	}
	row := r.data[r.idx]
	if len(dest) != len(row) {
		return fmt.Errorf("stub rows: expected %d destinations, got %d values", len(dest), len(row))
	}
	for i := range dest {
		if err := assignValue(row[i], dest[i]); err != nil {
			return fmt.Errorf("stub rows scan column %d: %w", i, err)
		}
	}
	return nil
}

func (r *StubRows) Close()                                       {}
func (r *StubRows) Err() error                                   { return r.err }
func (r *StubRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("SELECT 0") }
func (r *StubRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *StubRows) Values() ([]any, error)                       { return nil, nil }
func (r *StubRows) RawValues() [][]byte                          { return nil }
func (r *StubRows) Conn() *pgx.Conn                              { return nil }

// StubTx имитирует pgx.Tx — транзакцию. QueryRow/Exec маршрутизируются
// по подстроке SQL, как и в FakePgPool.
type StubTx struct {
	rows        map[string]*StubRow
	tags        map[string]pgconn.CommandTag
	commitErr   error
	rollbackErr error
}

// NewStubTx создаёт пустую транзакцию-заглушку.
func NewStubTx() *StubTx {
	return &StubTx{
		rows: make(map[string]*StubRow),
		tags: make(map[string]pgconn.CommandTag),
	}
}

// OnQueryRow регистрирует ответ для SQL, содержащего заданную подстроку.
func (t *StubTx) OnQueryRow(sqlContains string, row *StubRow) *StubTx {
	t.rows[sqlContains] = row
	return t
}

// OnExec регистрирует CommandTag для SQL, содержащего заданную подстроку.
func (t *StubTx) OnExec(sqlContains string, tag pgconn.CommandTag) *StubTx {
	t.tags[sqlContains] = tag
	return t
}

// OnCommit задаёт ошибку, которую вернёт Commit.
func (t *StubTx) OnCommit(err error) *StubTx {
	t.commitErr = err
	return t
}

func (t *StubTx) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	for key, row := range t.rows {
		if strings.Contains(sql, key) {
			return row
		}
	}
	return NewStubRow().WithError(pgx.ErrNoRows)
}

func (t *StubTx) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	for key, tag := range t.tags {
		if strings.Contains(sql, key) {
			return tag, nil
		}
	}
	return pgconn.NewCommandTag(""), nil
}

func (t *StubTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return NewStubRows(), nil
}

func (t *StubTx) Begin(_ context.Context) (pgx.Tx, error) { return t, nil }
func (t *StubTx) Commit(_ context.Context) error          { return t.commitErr }
func (t *StubTx) Rollback(_ context.Context) error        { return t.rollbackErr }
func (t *StubTx) Conn() *pgx.Conn                         { return nil }
func (t *StubTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *StubTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults { return nil }
func (t *StubTx) LargeObjects() pgx.LargeObjects                             { return pgx.LargeObjects{} }
func (t *StubTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

// FakePgPool — «замоканный PostgreSQL»: реализует service.PgPool и отвечает
// на SQL-запросы заранее зарегистрированными строками/наборами строк/тегами.
type FakePgPool struct {
	tx   *StubTx
	rows map[string]*StubRows
	row  map[string]*StubRow
	tags map[string]pgconn.CommandTag

	beginErr   error
	queryErr   error
	execErr    error
	queryRowFn func(sql string, args []any) pgx.Row
}

// NewFakePgPool создаёт пустой «замоканный PostgreSQL».
func NewFakePgPool() *FakePgPool {
	return &FakePgPool{
		tx:   NewStubTx(),
		rows: make(map[string]*StubRows),
		row:  make(map[string]*StubRow),
		tags: make(map[string]pgconn.CommandTag),
	}
}

// SetBeginTx задаёт транзакцию, возвращаемую методом Begin.
func (f *FakePgPool) SetBeginTx(tx *StubTx) *FakePgPool { f.tx = tx; return f }

// SetQueryRow регистрирует строку-ответ для SQL, содержащего подстроку.
func (f *FakePgPool) SetQueryRow(sqlContains string, row *StubRow) *FakePgPool {
	f.row[sqlContains] = row
	return f
}

// SetQuery регистрирует набор строк-ответ для SQL, содержащего подстроку.
func (f *FakePgPool) SetQuery(sqlContains string, rows *StubRows) *FakePgPool {
	f.rows[sqlContains] = rows
	return f
}

// SetExec регистрирует CommandTag для SQL, содержащего подстроку.
func (f *FakePgPool) SetExec(sqlContains string, tag pgconn.CommandTag) *FakePgPool {
	f.tags[sqlContains] = tag
	return f
}

// SetQueryRowFn задаёт функцию маршрутизации QueryRow: она получает SQL и
// аргументы и возвращает строку-ответ. Удобно, когда один и тот же SQL
// обслуживает несколько наборов данных (например, Sub() для разных sub_id).
func (f *FakePgPool) SetQueryRowFn(fn func(sql string, args []any) pgx.Row) *FakePgPool {
	f.queryRowFn = fn
	return f
}

// SetBeginErr задаёт ошибку для Begin.
func (f *FakePgPool) SetBeginErr(err error) *FakePgPool { f.beginErr = err; return f }

// SetQueryErr задаёт ошибку для Query/QueryRow/Exec.
func (f *FakePgPool) SetQueryErr(err error) *FakePgPool { f.queryErr = err; return f }

// SetExecErr задаёт ошибку для Exec.
func (f *FakePgPool) SetExecErr(err error) *FakePgPool { f.execErr = err; return f }

func (f *FakePgPool) Begin(_ context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

func (f *FakePgPool) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	for key, rows := range f.rows {
		if strings.Contains(sql, key) {
			return rows, nil
		}
	}
	return NewStubRows(), nil
}

func (f *FakePgPool) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	if f.queryRowFn != nil {
		return f.queryRowFn(sql, args)
	}
	for key, row := range f.row {
		if strings.Contains(sql, key) {
			return row
		}
	}
	return NewStubRow().WithError(pgx.ErrNoRows)
}

func (f *FakePgPool) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	for key, tag := range f.tags {
		if strings.Contains(sql, key) {
			return tag, nil
		}
	}
	return pgconn.NewCommandTag(""), nil
}

// Tx возвращает текущую транзакцию-заглушку (удобно для настройки tx-ответов).
func (f *FakePgPool) Tx() *StubTx { return f.tx }

// assignValue присваивает значение value в dest (указатель). Поддерживаются
// примитивы, time.Time, uuid.UUID, pgtype-типы и именованные int-типы.
func assignValue(value any, dest any) error {
	if value == nil {
		// nil в dest-указатель не пишем — оставляем нулевое значение.
		return nil
	}
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return errors.New("destination must be a non-nil pointer")
	}
	ev := dv.Elem()

	switch d := dest.(type) {
	case *int:
		*d = asInt(value)
		return nil
	case *int32:
		*d = int32(asInt(value))
		return nil
	case *int64:
		*d = int64(asInt(value))
		return nil
	case *string:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot assign %T to *string", value)
		}
		*d = s
		return nil
	case *bool:
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot assign %T to *bool", value)
		}
		*d = b
		return nil
	case *float64:
		fl, ok := value.(float64)
		if !ok {
			return fmt.Errorf("cannot assign %T to *float64", value)
		}
		*d = fl
		return nil
	case *time.Time:
		tm, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("cannot assign %T to *time.Time", value)
		}
		*d = tm
		return nil
	case *uuid.UUID:
		id, ok := value.(uuid.UUID)
		if !ok {
			return fmt.Errorf("cannot assign %T to *uuid.UUID", value)
		}
		*d = id
		return nil
	case **string:
		if value == nil {
			*d = nil
			return nil
		}
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot assign %T to **string", value)
		}
		*d = &s
		return nil
	case *pgtype.Date:
		tm, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("cannot assign %T to *pgtype.Date", value)
		}
		d.Time = tm
		d.Valid = true
		return nil
	case *pgtype.Int4:
		d.Int32 = int32(asInt(value))
		d.Valid = true
		return nil
	case *pgtype.Timestamp:
		tm, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("cannot assign %T to *pgtype.Timestamp", value)
		}
		d.Time = tm
		d.Valid = true
		return nil
	}

	// Именованные целочисленные типы (domain.SubID, domain.PlanID и т.п.).
	if ev.Kind() == reflect.Int || ev.Kind() == reflect.Int32 || ev.Kind() == reflect.Int64 {
		ev.SetInt(int64(asInt(value)))
		return nil
	}
	if ev.Kind() == reflect.Uint || ev.Kind() == reflect.Uint32 || ev.Kind() == reflect.Uint64 {
		ev.SetUint(uint64(asInt(value)))
		return nil
	}
	// Именованные строковые типы (domain.Role и т.п.).
	if ev.Kind() == reflect.String {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot assign %T to string-based type %T", value, dest)
		}
		ev.SetString(s)
		return nil
	}
	return fmt.Errorf("unsupported destination type %T", dest)
}

func asInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case uint:
		return int(x)
	case uint32:
		return int(x)
	case uint64:
		return int(x)
	}
	// Именованные целочисленные типы (domain.SubID, domain.PlanID и т.п.).
	rv := reflect.ValueOf(v)
	if rv.IsValid() {
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(rv.Uint())
		}
	}
	return 0
}

// CommandTagDelete создаёт CommandTag с заданным числом затронутых строк.
func CommandTagDelete(rowsAffected int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", rowsAffected))
}

// CommandTagInsert создаёт CommandTag вида "INSERT 0 N".
func CommandTagInsert(rowsAffected int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", rowsAffected))
}

// CommandTagUpdate создаёт CommandTag "UPDATE N".
func CommandTagUpdate(rowsAffected int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", rowsAffected))
}
