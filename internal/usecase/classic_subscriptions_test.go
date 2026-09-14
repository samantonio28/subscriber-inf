package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// ---------------------------------------------------------------------------
// Классические тесты: PostgreSQL НЕ поднимается. Вместо этого настоящим
// репозиториям (service.NewSubRepo и т.д.) передаётся FakePgPool —
// «замоканный коннект к БД». Репозитории «думают», что общаются с реальным
// PostgreSQL, хотя все строки/теги приходят из заглушек testutil/pgxstubs.go.
// ---------------------------------------------------------------------------

// субId, возвращаемый INSERT ... RETURNING sub_id
func stubSubRow(id domain.SubID, userID interface{}, planID, price int, subType string, start, end time.Time) *testutil.StubRow {
	return testutil.NewStubRow(
		id,        // s.sub_id        -> domain.SubID
		userID,    // s.user_id       -> uuid.UUID
		planID,    // s.plan_id       -> int
		nil,       // promocode_id    -> pgtype.Int4 (NULL)
		price,     // s.price         -> int
		subType,   // s.sub_type      -> string
		start,     // s.start_date    -> time.Time
		end,       // s.end_date      -> pgtype.Date (может быть zero/nil)
		"Netflix", // sv.service_name -> string
	)
}

func TestCreateSubUC_NewSub_Classic(t *testing.T) {
	ctx := context.Background()

	// Arrange: «замоканный PostgreSQL» отвечает на INSERT подписки.
	fakePG := testutil.NewFakePgPool()
	fakePG.Tx().OnQueryRow("RETURNING sub_id", testutil.NewStubRow(42))

	subRepo, err := service.NewSubRepo(fakePG)
	if err != nil {
		t.Fatalf("failed to create real SubRepo on fake pool: %v", err)
	}
	uc, err := NewCreateSubUC(subRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	now := time.Now()
	input := SubscriptionDTO{
		UserId:      testutil.TestUUID(),
		ServiceName: "Netflix",
		Price:       1000,
		SubType:     "usual",
		StartDate:   now,
		EndDate:     now.AddDate(0, 1, 0),
		PlanID:      1,
	}

	// Act
	got, err := uc.NewSub(ctx, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != 42 {
		t.Errorf("expected sub id 42, got %d", got)
	}
}

func TestGetSubUC_SubById_Classic(t *testing.T) {
	ctx := context.Background()
	userID := testutil.TestUUID()
	start := time.Now()
	end := time.Now().AddDate(0, 1, 0)

	// Arrange
	fakePG := testutil.NewFakePgPool()
	fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(7, userID, 1, 1000, "usual", start, end))

	subRepo, _ := service.NewSubRepo(fakePG)
	uc, err := NewGetSubUC(subRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	// Act
	dto, err := uc.SubById(ctx, 7)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dto.SubId != 7 || dto.ServiceName != "Netflix" || dto.Price != 1000 {
		t.Errorf("unexpected DTO: %+v", dto)
	}
	if dto.UserId != userID {
		t.Errorf("expected user id %v, got %v", userID, dto.UserId)
	}
}

func TestGetSubsUC_SubsByUserId_Classic(t *testing.T) {
	ctx := context.Background()
	userID := testutil.TestUUID()
	start := time.Now()
	end := time.Now().AddDate(0, 1, 0)

	// Arrange: SELECT sub_id возвращает [1], а Sub(1) — полную подписку.
	fakePG := testutil.NewFakePgPool()
	fakePG.SetQuery("SELECT sub_id FROM subscriptions", testutil.NewStubRows([]any{1}))
	fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 500, "usual", start, end))

	subRepo, _ := service.NewSubRepo(fakePG)
	uc, err := NewGetSubsUC(subRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	// Act
	subs, err := uc.SubsByUserId(ctx, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].SubId != 1 || subs[0].Price != 500 {
		t.Errorf("unexpected subscription: %+v", subs[0])
	}
}

func TestDeleteSubUC_DeleteSub_Classic(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes existing subscription", func(t *testing.T) {
		// Arrange
		fakePG := testutil.NewFakePgPool()
		fakePG.SetExec("DELETE FROM subscriptions", testutil.CommandTagDelete(1))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, err := NewDeleteSubUC(subRepo, &testutil.NopLogger{})
		if err != nil {
			t.Fatalf("failed to create usecase: %v", err)
		}

		// Act
		err = uc.DeleteSub(ctx, 5)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing subscription returns domain error", func(t *testing.T) {
		// Arrange
		fakePG := testutil.NewFakePgPool()
		fakePG.SetExec("DELETE FROM subscriptions", testutil.CommandTagDelete(0))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewDeleteSubUC(subRepo, &testutil.NopLogger{})

		// Act
		err := uc.DeleteSub(ctx, 404)

		// Assert
		if !errors.Is(err, domain.ErrNoSubscriptionDeleted) {
			t.Fatalf("expected ErrNoSubscriptionDeleted, got %v", err)
		}
	})
}

func TestUpdateSubUC_UpdateSub_Classic(t *testing.T) {
	ctx := context.Background()
	userID := testutil.TestUUID()
	firstOfMonth := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := firstOfMonth.AddDate(0, 1, 0)

	// Arrange: текущая подписка в «БД» + UPDATE успешен.
	fakePG := testutil.NewFakePgPool()
	fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 1000, "usual", firstOfMonth, end))
	fakePG.Tx().OnExec("UPDATE subscriptions", testutil.CommandTagUpdate(1))

	subRepo, _ := service.NewSubRepo(fakePG)
	uc, err := NewUpdateSubUC(subRepo, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	input := SubscriptionDTO{
		UserId:      userID,
		ServiceName: "Netflix",
		Price:       1100,
		SubType:     "usual",
		StartDate:   firstOfMonth,
		EndDate:     end,
		PlanID:      1,
	}

	// Act
	err = uc.UpdateSub(ctx, 1, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
