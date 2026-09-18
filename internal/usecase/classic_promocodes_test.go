package usecase

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// Классические тесты: настоящие PromocodeRepo/SubRepo поверх «замоканного
// PostgreSQL» (FakePgPool из internal/testutil/pgxstubs.go).

func stubPromoRow(id int, value string, discount, maxUses int, status string) *testutil.StubRow {
	return testutil.NewStubRow(
		id,                          // promocode_id
		1,                           // service_id
		value,                       // promocode
		nil,                         // plan_id      (pgtype.Int4, NULL)
		nil,                         // sub_id       (pgtype.Int4, NULL)
		time.Now().AddDate(0, 0, 7), // expires_at (pgtype.Date)
		time.Now(),                  // created_at   (pgtype.Timestamp)
		discount,                    // discount
		maxUses,                     // max_uses
		0,                           // cur_uses
		status,                      // status
		30,                          // duration_days
	)
}

func TestCreatePromocodeUC_Create_Classic(t *testing.T) {
	ctx := context.Background()

	// Arrange
	fakePG := testutil.NewFakePgPool()
	fakePG.Tx().OnQueryRow("RETURNING promocode_id", testutil.NewStubRow(123))

	promoRepo, err := service.NewPromocodeRepo(fakePG)
	if err != nil {
		t.Fatalf("failed to create real PromocodeRepo: %v", err)
	}
	uc, err := NewCreatePromocodeUC(promoRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	input := CreatePromocodeInput{
		ServiceID:    1,
		Value:        "SUMMER2026",
		Discount:     20,
		MaxUses:      100,
		DurationDays: 7,
	}

	// Act
	id, err := uc.Create(ctx, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != domain.PromocodeID(123) {
		t.Errorf("expected promocode id 123, got %d", id)
	}
}

func TestGetPromocodeUC_ByID_Classic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	// Arrange: кэш — shared-зависимость (Redis), заменяется моком; ByID
	// идёт напрямую в репозиторий, поэтому ожиданий на кэш нет.
	fakePG := testutil.NewFakePgPool()
	fakePG.SetQueryRow("WHERE promocode_id = $1", stubPromoRow(7, "SUMMER", 20, 100, "ACTIVE"))

	promoRepo, _ := service.NewPromocodeRepo(fakePG)
	cache := mock.NewMockPromocodeCache(ctrl)
	uc, err := NewGetPromocodeUC(promoRepo, cache, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	// Act
	pc, err := uc.ByID(ctx, domain.PromocodeID(7))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pc.PromocodeID != 7 || pc.Value != "SUMMER" || pc.Discount != 20 {
		t.Errorf("unexpected promocode: %+v", pc)
	}
	if pc.Status != domain.PromocodeStatusActive {
		t.Errorf("expected ACTIVE status, got %q", pc.Status)
	}
}

func TestApplyPromocodeUC_Apply_Classic(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Arrange: подписка «в БД» — новая (start в будущем) И даты — 1-е числа,
	// потому что SubRepo.UpdateSub требует день = 1 для start/end.
	firstOfNextMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	subID := domain.SubID(1)
	userID := testutil.TestUUID()
	subRow := stubSubRow(subID, userID, 1, 1000, "usual", firstOfNextMonth, firstOfNextMonth.AddDate(0, 1, 0))
	promoRow := stubPromoRow(5, "SUMMER", 20, 100, "ACTIVE")

	fakePG := testutil.NewFakePgPool()
	fakePG.Tx().OnQueryRow("SELECT s.sub_id", subRow)    // Sub()
	fakePG.SetQueryRow("WHERE promocode = $1", promoRow) // GetByCode()
	fakePG.Tx().OnExec("UPDATE subscriptions", testutil.CommandTagUpdate(1))
	fakePG.SetExec("UPDATE promocodes SET", testutil.CommandTagUpdate(1)) // IncrementUses()

	subRepo, _ := service.NewSubRepo(fakePG)
	promoRepo, _ := service.NewPromocodeRepo(fakePG)
	uc, err := NewApplyPromocodeUC(subRepo, promoRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	// Act
	out, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(subID), PromocodeValue: "SUMMER"})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.DiscountApplied != 20 || out.NewPrice != 800 {
		t.Errorf("unexpected output: %+v", out)
	}
}
