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
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewGetPromocodeUC(mockRepo, nil)
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
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewGetPromocodeUC(mockRepo, mockLogger)
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
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)

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
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewGetPromocodeUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval by code", func(t *testing.T) {
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

		mockRepo.EXPECT().GetByCode(gomock.Any(), code).Return(expectedPromo, nil)

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

		mockRepo.EXPECT().GetByCode(gomock.Any(), code).Return(domain.Promocode{}, expectedErr)
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)

		promo, err := uc.ByCode(context.Background(), code)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if promo.Value != "" {
			t.Errorf("expected empty promocode, got %+v", promo)
		}
	})
}