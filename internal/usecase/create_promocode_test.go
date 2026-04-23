package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewCreatePromocodeUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewCreatePromocodeUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewCreatePromocodeUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewCreatePromocodeUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestCreatePromocodeUC_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls to avoid test failures due to unexpected logs
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	uc, err := NewCreatePromocodeUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		input := CreatePromocodeInput{
			ServiceID:    1,
			Value:        "SUMMER2025",
			PlanID:       nil,
			SubID:        nil,
			ExpiresAt:    time.Now().AddDate(0, 0, 7),
			Discount:     20,
			MaxUses:      100,
			DurationDays: 7,
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.PromocodeID(123), nil)
		
		id, err := uc.Create(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if id != 123 {
			t.Errorf("expected promocode ID 123, got %d", id)
		}
	})

	t.Run("invalid discount below zero", func(t *testing.T) {
		input := CreatePromocodeInput{
			ServiceID:    1,
			Value:        "TEST",
			Discount:     -5,
			MaxUses:      1,
			DurationDays: 3,
		}

		id, err := uc.Create(context.Background(), input)
		if err != domain.ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})

	t.Run("invalid discount above 100", func(t *testing.T) {
		input := CreatePromocodeInput{
			ServiceID:    1,
			Value:        "TEST",
			Discount:     150,
			MaxUses:      1,
			DurationDays: 3,
		}

		id, err := uc.Create(context.Background(), input)
		if err != domain.ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		input := CreatePromocodeInput{
			ServiceID:    1,
			Value:        "ERROR",
			Discount:     10,
			MaxUses:      1,
			DurationDays: 3,
		}

		expectedErr := domain.ErrPromocodeNotFound
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.PromocodeID(0), expectedErr)

		id, err := uc.Create(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})
}