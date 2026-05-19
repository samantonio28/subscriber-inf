package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewGetPromocodeUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockCache := mock.NewMockPromocodeCache(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(mockRepo, mockCache, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(nil, mockCache, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil cache", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(mockRepo, nil, mockLogger)
		if err != domain.ErrInvalidCache {
			t.Errorf("expected ErrInvalidCache, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(mockRepo, mockCache, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestGetPromocodeUC_ByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockCache := mock.NewMockPromocodeCache(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).AnyTimes()

	uc, err := NewGetPromocodeUC(mockRepo, mockCache, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval", func(t *testing.T) {
		id := domain.PromocodeID(123)
		expectedPromo := domain.Promocode{
			PromocodeID:  id,
			ServiceID:    1,
			Value:        "TEST",
			PlanID:       nil,
			SubID:        nil,
			ExpiresAt:    time.Now().AddDate(0, 0, 7),
			CreatedAt:    time.Now(),
			Discount:     10,
			MaxUses:      100,
			CurUses:      0,
			Status:       domain.PromocodeStatusActive,
			DurationDays: 7,
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(expectedPromo, nil)

		promo, err := uc.ByID(context.Background(), id)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if promo.PromocodeID != expectedPromo.PromocodeID {
			t.Errorf("expected promocode ID %d, got %d", expectedPromo.PromocodeID, promo.PromocodeID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		id := domain.PromocodeID(999)
		expectedErr := domain.ErrPromocodeNotFound

		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(domain.Promocode{}, expectedErr)

		promo, err := uc.ByID(context.Background(), id)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if promo.PromocodeID != 0 {
			t.Errorf("expected empty promocode, got %+v", promo)
		}
	})
}

func TestGetPromocodeUC_ByCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockCache := mock.NewMockPromocodeCache(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).AnyTimes()

	uc, err := NewGetPromocodeUC(mockRepo, mockCache, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval by code with cache miss", func(t *testing.T) {
		code := "SUMMER25"
		expectedPromo := domain.Promocode{
			PromocodeID:  456,
			ServiceID:    1,
			Value:        code,
			PlanID:       nil,
			SubID:        nil,
			ExpiresAt:    time.Now().AddDate(0, 0, 7),
			CreatedAt:    time.Now(),
			Discount:     15,
			MaxUses:      50,
			CurUses:      0,
			Status:       domain.PromocodeStatusActive,
			DurationDays: 7,
		}

		// Кэш не содержит промокод
		mockCache.EXPECT().GetPromocode(gomock.Any(), code).Return(domain.Promocode{}, domain.ErrInvalidCache)
		mockRepo.EXPECT().GetByCode(gomock.Any(), code).Return(expectedPromo, nil)
		mockCache.EXPECT().SetPromocode(gomock.Any(), code, expectedPromo, gomock.Any()).Return(nil)
		mockLogger.EXPECT().Info("promocode retrieved from db and cached", "code", code).Times(1)

		promo, err := uc.ByCode(context.Background(), code)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if promo.Value != expectedPromo.Value {
			t.Errorf("expected promocode value %s, got %s", expectedPromo.Value, promo.Value)
		}
	})

	t.Run("successful retrieval by code with cache hit", func(t *testing.T) {
		code := "WINTER30"
		expectedPromo := domain.Promocode{
			PromocodeID:  789,
			ServiceID:    2,
			Value:        code,
			PlanID:       nil,
			SubID:        nil,
			ExpiresAt:    time.Now().AddDate(0, 0, 7),
			CreatedAt:    time.Now(),
			Discount:     20,
			MaxUses:      100,
			CurUses:      0,
			Status:       domain.PromocodeStatusActive,
			DurationDays: 7,
		}

		// Кэш содержит промокод
		mockCache.EXPECT().GetPromocode(gomock.Any(), code).Return(expectedPromo, nil)
		mockLogger.EXPECT().Info("promocode retrieved from cache", "code", code).Times(1)
		// Репозиторий не вызывается
		// SetPromocode не вызывается

		promo, err := uc.ByCode(context.Background(), code)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if promo.Value != expectedPromo.Value {
			t.Errorf("expected promocode value %s, got %s", expectedPromo.Value, promo.Value)
		}
	})

	t.Run("not found by code", func(t *testing.T) {
		code := "INVALID"
		expectedErr := domain.ErrPromocodeNotFound

		// Кэш не содержит
		mockCache.EXPECT().GetPromocode(gomock.Any(), code).Return(domain.Promocode{}, domain.ErrInvalidCache)
		mockRepo.EXPECT().GetByCode(gomock.Any(), code).Return(domain.Promocode{}, expectedErr)
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

		promo, err := uc.ByCode(context.Background(), code)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if promo.Value != "" {
			t.Errorf("expected empty promocode, got %+v", promo)
		}
	})
}