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

func TestNewUpdateSubUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewUpdateSubUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewUpdateSubUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewUpdateSubUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestUpdateSubUC_UpdateSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewUpdateSubUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		subID := 123
		userID := uuid.New()
		existingSub := domain.Subscription{
			SubId:       domain.SubID(subID),
			UserID:      userID,
			ServiceName: "Old Service",
			Price:       1000,
			SubType:     domain.SubTypeUsual,
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
		}
		inputDTO := SubscriptionDTO{
			UserId:      userID,
			ServiceName: "New Service",
			Price:       1500,
			SubType:     "usual",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 2, 0),
		}

		mockLogger.EXPECT().Info("Updating subscription", subID).Times(1)
		mockRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(existingSub, nil)
		mockRepo.EXPECT().UpdateSub(gomock.Any(), gomock.Any()).Return(nil)
		mockLogger.EXPECT().Info("subscription updated:", subID).Times(1)

		err := uc.UpdateSub(context.Background(), subID, inputDTO)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("subscription not found", func(t *testing.T) {
		subID := 999
		inputDTO := SubscriptionDTO{}

		mockLogger.EXPECT().Info("Updating subscription", subID).Times(1)
		mockRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound)
		mockLogger.EXPECT().Error("not exists:", subID, domain.ErrSubscriptionNotFound).Times(1)

		err := uc.UpdateSub(context.Background(), subID, inputDTO)
		if err != domain.ErrSubscriptionNotFound {
			t.Errorf("expected error %v, got %v", domain.ErrSubscriptionNotFound, err)
		}
	})

	t.Run("invalid DTO conversion", func(t *testing.T) {
		subID := 123
		userID := uuid.New()
		existingSub := domain.Subscription{
			SubId:       domain.SubID(subID),
			UserID:      userID,
			ServiceName: "Old Service",
			Price:       1000,
			SubType:     domain.SubTypeUsual,
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
		}
		// Invalid SubType
		inputDTO := SubscriptionDTO{
			UserId:      userID,
			ServiceName: "New Service",
			Price:       1500,
			SubType:     "invalid",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 2, 0),
		}

		mockLogger.EXPECT().Info("Updating subscription", subID).Times(1)
		mockRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(existingSub, nil)
		mockLogger.EXPECT().Error("invalid input:", inputDTO, gomock.Any()).Times(1)

		err := uc.UpdateSub(context.Background(), subID, inputDTO)
		if err == nil {
			t.Error("expected error for invalid SubType")
		}
	})

	t.Run("repository update error", func(t *testing.T) {
		subID := 123
		userID := uuid.New()
		existingSub := domain.Subscription{
			SubId:       domain.SubID(subID),
			UserID:      userID,
			ServiceName: "Old Service",
			Price:       1000,
			SubType:     domain.SubTypeUsual,
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
		}
		inputDTO := SubscriptionDTO{
			UserId:      userID,
			ServiceName: "New Service",
			Price:       1500,
			SubType:     "usual",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 2, 0),
		}
		expectedErr := domain.ErrSubscriptionNotFound

		mockLogger.EXPECT().Info("Updating subscription", subID).Times(1)
		mockRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(existingSub, nil)
		mockRepo.EXPECT().UpdateSub(gomock.Any(), gomock.Any()).Return(expectedErr)
		mockLogger.EXPECT().Error("error updating subscription:", subID, expectedErr).Times(1)

		err := uc.UpdateSub(context.Background(), subID, inputDTO)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}