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

func TestNewTotalCostsUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewTotalCostsUC(mockRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		uc, err := NewTotalCostsUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewTotalCostsUC(mockRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestTotalCostsUC_TotalCosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	uc, err := NewTotalCostsUC(mockRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	t.Run("successful calculation", func(t *testing.T) {
		input := SubsFilterDTO{
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
			UserID:      uuid.New(),
			ServiceName: "Netflix",
			SubType:     "usual",
		}
		expectedSum := 5000
		expectedSubIds := []domain.SubID{1, 2, 3}

		mockLogger.EXPECT().Info("TotalCosts", "input", input).Times(1)
		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(expectedSum, expectedSubIds, nil)
		mockLogger.EXPECT().Info("TotalCosts", "input", input, "output", expectedSum, "subIds len", 3).Times(1)

		sum, subIds, err := uc.TotalCosts(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sum != expectedSum {
			t.Errorf("expected sum %d, got %d", expectedSum, sum)
		}
		if len(subIds) != len(expectedSubIds) {
			t.Errorf("expected %d sub IDs, got %d", len(expectedSubIds), len(subIds))
		}
	})

	t.Run("invalid filter conversion", func(t *testing.T) {
		input := SubsFilterDTO{
			StartDate:   time.Time{},
			EndDate:     time.Now(),
			UserID:      uuid.New(),
			ServiceName: "Netflix",
			SubType:     "invalid",
		}

		mockLogger.EXPECT().Info("TotalCosts", "input", input).Times(1)
		mockLogger.EXPECT().Error("TotalCosts", "input", input, "error", gomock.Any()).Times(1)

		sum, subIds, err := uc.TotalCosts(context.Background(), input)
		if err == nil {
			t.Error("expected error for invalid SubType")
		}
		if sum != 0 {
			t.Errorf("expected sum 0, got %d", sum)
		}
		if subIds != nil {
			t.Errorf("expected nil subIds, got %v", subIds)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		input := SubsFilterDTO{
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
			UserID:      uuid.New(),
			ServiceName: "Netflix",
			SubType:     "usual",
		}
		expectedErr := domain.ErrSubscriptionNotFound

		mockLogger.EXPECT().Info("TotalCosts", "input", input).Times(1)
		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(0, nil, expectedErr)
		mockLogger.EXPECT().Error("TotalCosts", "input", input, "error", expectedErr).Times(1)

		sum, subIds, err := uc.TotalCosts(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if sum != 0 {
			t.Errorf("expected sum 0, got %d", sum)
		}
		if subIds != nil {
			t.Errorf("expected nil subIds, got %v", subIds)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		input := SubsFilterDTO{
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
			UserID:      uuid.New(),
			ServiceName: "Nonexistent",
			SubType:     "usual",
		}
		expectedSum := 0
		expectedSubIds := []domain.SubID{}

		mockLogger.EXPECT().Info("TotalCosts", "input", input).Times(1)
		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(expectedSum, expectedSubIds, nil)
		mockLogger.EXPECT().Info("TotalCosts", "input", input, "output", expectedSum, "subIds len", 0).Times(1)

		sum, subIds, err := uc.TotalCosts(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sum != 0 {
			t.Errorf("expected sum 0, got %d", sum)
		}
		if len(subIds) != 0 {
			t.Errorf("expected 0 sub IDs, got %d", len(subIds))
		}
	})
}