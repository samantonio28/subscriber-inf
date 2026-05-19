package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewUpdateSubscriptionPlanUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockCache := mock.NewMockSubscriptionPlanCache(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewUpdateSubscriptionPlanUC(mockRepo, mockCache, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewUpdateSubscriptionPlanUC(nil, mockCache, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil cache", func(t *testing.T) {
		uc, err := NewUpdateSubscriptionPlanUC(mockRepo, nil, mockLogger)
		if err != domain.ErrInvalidCache {
			t.Errorf("expected ErrInvalidCache, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewUpdateSubscriptionPlanUC(mockRepo, mockCache, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestUpdateSubscriptionPlanUC_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockCache := mock.NewMockSubscriptionPlanCache(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewUpdateSubscriptionPlanUC(mockRepo, mockCache, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		input := UpdateSubscriptionPlanInput{
			ID:           domain.PlanID(123),
			ServiceID:    1,
			Name:         "Updated Plan",
			DurationDays: 30,
			Price:        2500,
		}
		cacheKey := "123"

		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockCache.EXPECT().DeleteSubscriptionPlan(gomock.Any(), cacheKey).Return(nil)
		mockLogger.EXPECT().Info("subscription plan updated").Times(1)

		err := uc.Update(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid input - zero service ID", func(t *testing.T) {
		input := UpdateSubscriptionPlanInput{
			ID:           domain.PlanID(123),
			ServiceID:    0,
			Name:         "Invalid",
			DurationDays: 7,
			Price:        999,
		}

		err := uc.Update(context.Background(), input)
		if err == nil {
			t.Error("expected error for zero service ID")
		}
	})

	t.Run("invalid input - negative price", func(t *testing.T) {
		input := UpdateSubscriptionPlanInput{
			ID:           domain.PlanID(123),
			ServiceID:    1,
			Name:         "Negative",
			DurationDays: 7,
			Price:        -100,
		}

		err := uc.Update(context.Background(), input)
		if err == nil {
			t.Error("expected error for negative price")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		input := UpdateSubscriptionPlanInput{
			ID:           domain.PlanID(456),
			ServiceID:    1,
			Name:         "Error",
			DurationDays: 7,
			Price:        999,
		}

		expectedErr := domain.ErrSubscriptionPlanNotFound
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(expectedErr)
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)
		// cache invalidation should NOT be called because update failed

		err := uc.Update(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}