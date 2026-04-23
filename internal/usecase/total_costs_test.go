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

	// Allow any logger calls to avoid test failures due to unexpected logs
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

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
		// Create dummy subscriptions
		expectedSubs := []domain.Subscription{
			{SubId: 1, UserID: input.UserID, ServiceName: "Netflix", Price: 2000, SubType: domain.SubTypeUsual, StartDate: input.StartDate, EndDate: input.EndDate},
			{SubId: 2, UserID: input.UserID, ServiceName: "Netflix", Price: 2000, SubType: domain.SubTypeUsual, StartDate: input.StartDate, EndDate: input.EndDate},
			{SubId: 3, UserID: input.UserID, ServiceName: "Netflix", Price: 1000, SubType: domain.SubTypeUsual, StartDate: input.StartDate, EndDate: input.EndDate},
		}

		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(expectedSum, expectedSubs, nil)

		sum, subs, err := uc.TotalCosts(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sum != expectedSum {
			t.Errorf("expected sum %d, got %d", expectedSum, sum)
		}
		if len(subs) != len(expectedSubs) {
			t.Errorf("expected %d subs, got %d", len(expectedSubs), len(subs))
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

		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(0, nil, expectedErr)

		sum, subs, err := uc.TotalCosts(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if sum != 0 {
			t.Errorf("expected sum 0, got %d", sum)
		}
		if subs != nil {
			t.Errorf("expected nil subs, got %v", subs)
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
		expectedSubs := []domain.Subscription{}

		mockRepo.EXPECT().SubsTotalCosts(gomock.Any(), gomock.Any()).Return(expectedSum, expectedSubs, nil)

		sum, subs, err := uc.TotalCosts(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sum != 0 {
			t.Errorf("expected sum 0, got %d", sum)
		}
		if len(subs) != 0 {
			t.Errorf("expected 0 subs, got %d", len(subs))
		}
	})
}