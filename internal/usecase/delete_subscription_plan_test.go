package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

func TestNewDeleteSubscriptionPlanUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockCache := mock.NewMockSubscriptionPlanCache(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewDeleteSubscriptionPlanUC(mockRepo, mockCache, &testutil.NopLogger{})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewDeleteSubscriptionPlanUC(nil, mockCache, &testutil.NopLogger{})
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil cache", func(t *testing.T) {
		uc, err := NewDeleteSubscriptionPlanUC(mockRepo, nil, &testutil.NopLogger{})
		if err != domain.ErrInvalidCache {
			t.Errorf("expected ErrInvalidCache, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewDeleteSubscriptionPlanUC(mockRepo, mockCache, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestDeleteSubscriptionPlanUC_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockCache := mock.NewMockSubscriptionPlanCache(ctrl)

	uc, err := NewDeleteSubscriptionPlanUC(mockRepo, mockCache, &testutil.NopLogger{})
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful deletion", func(t *testing.T) {
		id := domain.PlanID(123)
		cacheKey := "123"

		mockRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)
		mockCache.EXPECT().DeleteSubscriptionPlan(gomock.Any(), cacheKey).Return(nil)

		err := uc.Delete(context.Background(), id)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		id := domain.PlanID(456)
		expectedErr := domain.ErrSubscriptionPlanNotFound

		mockRepo.EXPECT().Delete(gomock.Any(), id).Return(expectedErr)
		// cache invalidation should NOT be called because deletion failed

		err := uc.Delete(context.Background(), id)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
