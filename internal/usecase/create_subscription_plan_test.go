package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewCreateSubscriptionPlanUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewCreateSubscriptionPlanUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewCreateSubscriptionPlanUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewCreateSubscriptionPlanUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestCreateSubscriptionPlanUC_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewCreateSubscriptionPlanUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		input := CreateSubscriptionPlanInput{
			ServiceID:    1,
			Name:         "Premium",
			DurationDays: 30,
			Price:        2999,
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.PlanID(123), nil)
		mockLogger.EXPECT().Info("subscription plan created").Times(1)

		id, err := uc.Create(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if id != 123 {
			t.Errorf("expected plan ID 123, got %d", id)
		}
	})

	t.Run("invalid input - zero service ID", func(t *testing.T) {
		input := CreateSubscriptionPlanInput{
			ServiceID:    0,
			Name:         "Invalid",
			DurationDays: 7,
			Price:        999,
		}

		id, err := uc.Create(context.Background(), input)
		if err == nil {
			t.Error("expected error for zero service ID")
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})

	t.Run("invalid input - negative price", func(t *testing.T) {
		input := CreateSubscriptionPlanInput{
			ServiceID:    1,
			Name:         "Negative",
			DurationDays: 7,
			Price:        -100,
		}

		id, err := uc.Create(context.Background(), input)
		if err == nil {
			t.Error("expected error for negative price")
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		input := CreateSubscriptionPlanInput{
			ServiceID:    1,
			Name:         "Error",
			DurationDays: 7,
			Price:        999,
		}

		expectedErr := domain.ErrSubscriptionPlanNotFound
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.PlanID(0), expectedErr)
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)

		id, err := uc.Create(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if id != 0 {
			t.Errorf("expected ID 0, got %d", id)
		}
	})
}