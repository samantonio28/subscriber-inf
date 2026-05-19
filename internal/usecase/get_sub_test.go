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

func TestNewGetSubUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewGetSubUC(mockSubRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil subscription repository", func(t *testing.T) {
		uc, err := NewGetSubUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewGetSubUC(mockSubRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestGetSubUC_SubById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewGetSubUC(mockSubRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	validUserID := uuid.New()
	validStartDate := time.Now()
	validEndDate := validStartDate.AddDate(0, 1, 0)

	t.Run("successful retrieval", func(t *testing.T) {
		subID := 42
		expectedSub := domain.Subscription{
			SubId:       domain.SubID(subID),
			UserID:      validUserID,
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     domain.SubTypeUsual,
			StartDate:   validStartDate,
			EndDate:     validEndDate,
		}
		expectedDTO := SubToDTO(expectedSub)

		mockLogger.EXPECT().Info("getting subscription by id", subID).Times(1)
		mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(expectedSub, nil)
		mockLogger.EXPECT().Info("got subscription by id", subID, ": ", expectedSub).Times(1)

		dto, err := uc.SubById(context.Background(), subID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dto != expectedDTO {
			t.Errorf("expected DTO %+v, got %+v", expectedDTO, dto)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		subID := 99
		expectedErr := domain.ErrSubscriptionNotFound

		mockLogger.EXPECT().Info("getting subscription by id", subID).Times(1)
		mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(subID)).Return(domain.Subscription{}, expectedErr)
		mockLogger.EXPECT().Error("error getting subscription by id", subID, expectedErr).Times(1)

		dto, err := uc.SubById(context.Background(), subID)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if dto != (SubscriptionDTO{}) {
			t.Errorf("expected empty DTO, got %+v", dto)
		}
	})
}
