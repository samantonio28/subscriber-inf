package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// ---------------------------------------------------------------------------
// Лондонский стиль: репозитории заменены gomock-моками, проверяется
// взаимодействие use case с коллабораторами (вызовы и их аргументы).
// ---------------------------------------------------------------------------

func TestNewApplyPromocodeUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSub := mock.NewMockSubscriptionRepository(ctrl)
	mockPromo := mock.NewMockPromocodeRepository(ctrl)
	nop := &testutil.NopLogger{}

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewApplyPromocodeUC(mockSub, mockPromo, nop)
		if err != nil || uc == nil {
			t.Fatalf("expected ok, got uc=%v err=%v", uc, err)
		}
	})
	t.Run("nil sub repository", func(t *testing.T) {
		uc, err := NewApplyPromocodeUC(nil, mockPromo, nop)
		if !errors.Is(err, domain.ErrInvalidSubRepo) || uc != nil {
			t.Fatalf("expected ErrInvalidSubRepo, got %v", err)
		}
	})
	t.Run("nil promocode repository", func(t *testing.T) {
		uc, err := NewApplyPromocodeUC(mockSub, nil, nop)
		if !errors.Is(err, domain.ErrInvalidSubRepo) || uc != nil {
			t.Fatalf("expected ErrInvalidSubRepo, got %v", err)
		}
	})
	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewApplyPromocodeUC(mockSub, mockPromo, nil)
		if !errors.Is(err, domain.ErrInvalidLogger) || uc != nil {
			t.Fatalf("expected ErrInvalidLogger, got %v", err)
		}
	})
}

func TestApplyPromocodeUC_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	now := time.Now()

	t.Run("applies discount to a new subscription", func(t *testing.T) {
		// Arrange
		sub := domain.Subscription{
			SubId:       1,
			UserID:      testutil.TestUUID(),
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     domain.SubTypeUsual,
			StartDate:   now.Add(time.Hour), // будущая — «новая» подписка
			EndDate:     now.Add(30 * 24 * time.Hour),
			PlanID:      1,
		}
		promo := domain.Promocode{
			PromocodeID: 5,
			ServiceID:   1,
			Value:       "SUMMER",
			ExpiresAt:   now.Add(7 * 24 * time.Hour),
			Discount:    20,
			MaxUses:     100,
			Status:      domain.PromocodeStatusActive,
		}

		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)

		mockSub.EXPECT().Sub(ctx, sub.SubId).Return(sub, nil)
		mockPromo.EXPECT().GetByCode(ctx, promo.Value).Return(promo, nil)
		mockSub.EXPECT().UpdateSub(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, s domain.Subscription) error {
				if s.Price != 800 {
					t.Errorf("expected updated price 800, got %d", s.Price)
				}
				if s.SubType.String() != "promocode" {
					t.Errorf("expected sub_type promocode, got %q", s.SubType.String())
				}
				return nil
			},
		)
		mockPromo.EXPECT().IncrementUses(ctx, promo.PromocodeID).Return(nil)

		uc, err := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})
		if err != nil {
			t.Fatalf("failed to create usecase: %v", err)
		}

		// Act
		out, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(sub.SubId), PromocodeValue: promo.Value})

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.DiscountApplied != 20 || out.NewPrice != 800 {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)

		mockSub.EXPECT().Sub(ctx, gomock.Any()).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound)

		uc, _ := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})

		out, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: 999, PromocodeValue: "X"})
		if !errors.Is(err, domain.ErrSubscriptionNotFound) || out != nil {
			t.Fatalf("expected ErrSubscriptionNotFound, got err=%v out=%+v", err, out)
		}
	})

	t.Run("inactive promocode rejected", func(t *testing.T) {
		sub := domain.Subscription{
			SubId: 1, UserID: testutil.TestUUID(), ServiceName: "N", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: now.Add(-time.Hour),
			EndDate: now.Add(30 * 24 * time.Hour), PlanID: 1,
		}
		promo := domain.Promocode{
			PromocodeID: 5, ServiceID: 1, Value: "OLD", ExpiresAt: now.Add(time.Hour),
			Discount: 20, MaxUses: 10, Status: domain.PromocodeStatusDisabled,
		}

		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)

		mockSub.EXPECT().Sub(ctx, sub.SubId).Return(sub, nil)
		mockPromo.EXPECT().GetByCode(ctx, promo.Value).Return(promo, nil)

		uc, _ := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})

		_, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(sub.SubId), PromocodeValue: promo.Value})
		if !errors.Is(err, domain.ErrPromocodeNotActive) {
			t.Fatalf("expected ErrPromocodeNotActive, got %v", err)
		}
	})

	t.Run("expired promocode rejected", func(t *testing.T) {
		sub := domain.Subscription{
			SubId: 1, UserID: testutil.TestUUID(), ServiceName: "N", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: now.Add(-time.Hour),
			EndDate: now.Add(30 * 24 * time.Hour), PlanID: 1,
		}
		promo := domain.Promocode{
			PromocodeID: 5, ServiceID: 1, Value: "OLD", ExpiresAt: now.Add(-time.Hour),
			Discount: 20, MaxUses: 10, Status: domain.PromocodeStatusActive,
		}

		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)

		mockSub.EXPECT().Sub(ctx, sub.SubId).Return(sub, nil)
		mockPromo.EXPECT().GetByCode(ctx, promo.Value).Return(promo, nil)

		uc, _ := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})

		_, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(sub.SubId), PromocodeValue: promo.Value})
		if !errors.Is(err, domain.ErrPromocodeExpired) {
			t.Fatalf("expected ErrPromocodeExpired, got %v", err)
		}
	})

	t.Run("subscription update failure propagated", func(t *testing.T) {
		sub := domain.Subscription{
			SubId: 1, UserID: testutil.TestUUID(), ServiceName: "N", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: now.Add(time.Hour),
			EndDate: now.Add(30 * 24 * time.Hour), PlanID: 1,
		}
		promo := domain.Promocode{
			PromocodeID: 5, ServiceID: 1, Value: "NEW", ExpiresAt: now.Add(time.Hour),
			Discount: 20, MaxUses: 10, Status: domain.PromocodeStatusActive,
		}

		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)

		mockSub.EXPECT().Sub(ctx, sub.SubId).Return(sub, nil)
		mockPromo.EXPECT().GetByCode(ctx, promo.Value).Return(promo, nil)
		mockSub.EXPECT().UpdateSub(ctx, gomock.Any()).Return(errors.New("db failure"))

		uc, _ := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})

		_, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(sub.SubId), PromocodeValue: promo.Value})
		if err == nil || err.Error() != "db failure" {
			t.Fatalf("expected db failure, got %v", err)
		}
	})
}
