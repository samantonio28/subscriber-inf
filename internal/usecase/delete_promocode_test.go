package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewDeletePromocodeUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewDeletePromocodeUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewDeletePromocodeUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewDeletePromocodeUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestDeletePromocodeUC_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPromocodeRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	uc, err := NewDeletePromocodeUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful deletion", func(t *testing.T) {
		id := domain.PromocodeID(123)

		mockRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)

		err := uc.Delete(context.Background(), id)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		id := domain.PromocodeID(456)
		expectedErr := domain.ErrPromocodeNotFound

		mockRepo.EXPECT().Delete(gomock.Any(), id).Return(expectedErr)

		err := uc.Delete(context.Background(), id)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
