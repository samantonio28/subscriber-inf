package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewGetSubscriptionPlanUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewGetSubscriptionPlanUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewGetSubscriptionPlanUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewGetSubscriptionPlanUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestGetSubscriptionPlanUC_ByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewGetSubscriptionPlanUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval", func(t *testing.T) {
		id := domain.PlanID(123)
		expectedPlan := domain.SubscriptionPlan{
			PlanID:       id,
			ServiceID:    1,
			Name:         "Premium",
			DurationDays: 30,
			Price:        2999,
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(expectedPlan, nil)
		mockLogger.EXPECT().Info("subscription plan retrieved").Times(1)

		plan, err := uc.ByID(context.Background(), id)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if plan.PlanID != expectedPlan.PlanID {
			t.Errorf("expected plan ID %d, got %d", expectedPlan.PlanID, plan.PlanID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		id := domain.PlanID(999)
		expectedErr := domain.ErrSubscriptionPlanNotFound

		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(domain.SubscriptionPlan{}, expectedErr)
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)

		plan, err := uc.ByID(context.Background(), id)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if plan.PlanID != 0 {
			t.Errorf("expected empty plan, got %+v", plan)
		}
	})
}

func TestGetSubscriptionPlanUC_ByServiceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionPlanRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewGetSubscriptionPlanUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval by service", func(t *testing.T) {
		serviceID := 1
		expectedPlans := []domain.SubscriptionPlan{
			{
				PlanID:       1,
				ServiceID:    serviceID,
				Name:         "Basic",
				DurationDays: 7,
				Price:        999,
			},
			{
				PlanID:       2,
				ServiceID:    serviceID,
				Name:         "Premium",
				DurationDays: 30,
				Price:        2999,
			},
		}

		mockRepo.EXPECT().GetByService(gomock.Any(), serviceID).Return(expectedPlans, nil)
		mockLogger.EXPECT().Info("subscription plans retrieved by service").Times(1)

		plans, err := uc.ByServiceID(context.Background(), serviceID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(plans) != 2 {
			t.Errorf("expected 2 plans, got %d", len(plans))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		serviceID := 2
		expectedPlans := []domain.SubscriptionPlan{}

		mockRepo.EXPECT().GetByService(gomock.Any(), serviceID).Return(expectedPlans, nil)
		mockLogger.EXPECT().Info("subscription plans retrieved by service").Times(1)

		plans, err := uc.ByServiceID(context.Background(), serviceID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(plans) != 0 {
			t.Errorf("expected 0 plans, got %d", len(plans))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		serviceID := 3
		expectedErr := domain.ErrSubscriptionPlanNotFound

		mockRepo.EXPECT().GetByService(gomock.Any(), serviceID).Return(nil, expectedErr)
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(nil)

		plans, err := uc.ByServiceID(context.Background(), serviceID)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if plans != nil {
			t.Errorf("expected nil plans, got %v", plans)
		}
	})
}