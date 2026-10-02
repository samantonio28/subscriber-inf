//go:build allure

package usecase

import (
	"os"
	"testing"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// Фикстура пакета: «замоканный PostgreSQL» (FakePgPool), заполняемый один раз
// перед прогоном allure-тестов (см. TestMain). Собран с тегом allure, поэтому
// существует только при `go test -tags allure` и не влияет на обычные
// unit-тесты (classic_*.go и т.п.).

var testPG *testutil.FakePgPool

// TestMain настраивает фикстуру до запуска allure-тестов.
func TestMain(m *testing.M) {
	testPG = newSeededPG()
	os.Exit(m.Run())
}

// newSeededPG собирает FakePgPool, наполненный типовыми данными: одна подписка
// и один промокод, которые «БД» отдаёт по умолчанию. Строки строятся теми же
// хелперами stubSubRow/stubPromoRow, что и в остальных тестах.
func newSeededPG() *testutil.FakePgPool {
	p := testutil.NewFakePgPool()

	p.SetQueryRow("WHERE sub_id = $1",
		stubSubRow(1, testutil.TestUUID(), 1, 1000, "usual", time.Now(), time.Now().AddDate(0, 1, 0)))
	p.SetQueryRow("WHERE promocode_id = $1",
		stubPromoRow(1, "SUMMER", 20, 100, "ACTIVE"))

	return p
}

// resetFixtures возвращает фикстуру к исходному состоянию и отдаёт свежий
// заполненный FakePgPool. Вызывай в начале теста, чтобы получить чистую
// фикстуру, не «протёкшую» из предыдущего теста.
func resetFixtures() *testutil.FakePgPool {
	testPG = newSeededPG()
	return testPG
}
