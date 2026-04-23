package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewGetSubsUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewGetSubsUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewGetSubsUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewGetSubsUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestGetSubsUC_SubsByUserId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls for Error (used in repository error case)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc, err := NewGetSubsUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful retrieval", func(t *testing.T) {
		userID := uuid.New()
		subs := []domain.Subscription{
			{
				SubId:       1,
				UserID:      userID,
				ServiceName: "Service A",
				Price:       100,
				SubType:     domain.SubTypeUsual,
				StartDate:   time.Now(),
				EndDate:     time.Now().AddDate(0, 1, 0),
			},
			{
				SubId:       2,
				UserID:      userID,
				ServiceName: "Service B",
				Price:       200,
				SubType:     domain.SubTypePromocode,
				StartDate:   time.Now(),
				EndDate:     time.Now().AddDate(0, 0, 7),
			},
		}

		mockLogger.EXPECT().Info("getting subscriptions by user id", "user_id", userID).Times(1)
		mockRepo.EXPECT().UserSubs(gomock.Any(), userID).Return(subs, nil)
		mockLogger.EXPECT().Info("got subscriptions by user id", "user_id", userID, "count", 2).Times(1)

		result, err := uc.SubsByUserId(context.Background(), userID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 subscriptions, got %d", len(result))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		userID := uuid.New()
		expectedErr := domain.ErrSubscriptionNotFound

		mockLogger.EXPECT().Info("getting subscriptions by user id", "user_id", userID).Times(1)
		mockRepo.EXPECT().UserSubs(gomock.Any(), userID).Return(nil, expectedErr)
		// Error call is already covered by AnyTimes() expectation

		result, err := uc.SubsByUserId(context.Background(), userID)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		userID := uuid.New()
		subs := []domain.Subscription{}

		mockLogger.EXPECT().Info("getting subscriptions by user id", "user_id", userID).Times(1)
		mockRepo.EXPECT().UserSubs(gomock.Any(), userID).Return(subs, nil)
		mockLogger.EXPECT().Info("got subscriptions by user id", "user_id", userID, "count", 0).Times(1)

		result, err := uc.SubsByUserId(context.Background(), userID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 subscriptions, got %d", len(result))
		}
	})
}