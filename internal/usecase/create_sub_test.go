package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
)

func TestNewCreateSubUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewCreateSubUC(mockSubRepo, mockLogger)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if uc == nil {
			t.Error("expected uc not nil")
		}
	})

	t.Run("nil subscription repository", func(t *testing.T) {
		uc, err := NewCreateSubUC(nil, mockLogger)
		if err != domain.ErrInvalidSubRepo {
			t.Errorf("expected ErrInvalidSubRepo, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		uc, err := NewCreateSubUC(mockSubRepo, nil)
		if err != domain.ErrInvalidLogger {
			t.Errorf("expected ErrInvalidLogger, got %v", err)
		}
		if uc != nil {
			t.Error("expected uc nil")
		}
	})
}

func TestCreateSubUC_NewSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	mockLogger := mock.NewMockLogger(ctrl)

	// Allow any logger calls to avoid test failures due to unexpected logs
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).AnyTimes()

	uc, err := NewCreateSubUC(mockSubRepo, mockLogger)
	if err != nil {
		t.Fatalf("failed to create usecase: %v", err)
	}

	validUserID := uuid.New()
	validStartDate := time.Now()
	validEndDate := validStartDate.AddDate(0, 1, 0)

	t.Run("successful subscription creation", func(t *testing.T) {
		input := SubscriptionDTO{
			UserId:      validUserID,
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     "usual",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			PlanID:      1,
			PromocodeID: nil,
		}

		expectedSub, err := DTOToSub(input)
		if err != nil {
			t.Fatalf("failed to convert DTO: %v", err)
		}

		mockSubRepo.EXPECT().StoreSub(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, sub domain.Subscription) (domain.SubID, error) {
				if sub != expectedSub {
					t.Errorf("expected sub %+v, got %+v", expectedSub, sub)
				}
				return domain.SubID(42), nil
			},
		)

		subID, err := uc.NewSub(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if subID != 42 {
			t.Errorf("expected subID 42, got %d", subID)
		}
	})

	t.Run("successful subscription creation with nil user ID generates new UUID", func(t *testing.T) {
		input := SubscriptionDTO{
			UserId:      uuid.Nil,
			ServiceName: "Spotify",
			Price:       500,
			SubType:     "usual",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			PlanID:      2,
			PromocodeID: nil,
		}

		mockSubRepo.EXPECT().StoreSub(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, sub domain.Subscription) (domain.SubID, error) {
				if sub.UserID == uuid.Nil {
					t.Error("expected generated user ID, got nil")
				}
				return domain.SubID(43), nil
			},
		)

		subID, err := uc.NewSub(context.Background(), input)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if subID != 43 {
			t.Errorf("expected subID 43, got %d", subID)
		}
	})

	t.Run("invalid DTO leads to error", func(t *testing.T) {
		// Invalid SubType that will cause domain.NewSubscription to error
		input := SubscriptionDTO{
			UserId:      validUserID,
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     "invalid_type", // not a valid SubType enum
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			PlanID:      1,
			PromocodeID: nil,
		}

		// WithFields returns a logger.Entry, which is a *logrus.Entry.
		// We can return nil as a placeholder.

		subID, err := uc.NewSub(context.Background(), input)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if subID != 0 {
			t.Errorf("expected subID 0, got %d", subID)
		}
	})

	t.Run("repository store error", func(t *testing.T) {
		input := SubscriptionDTO{
			UserId:      validUserID,
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     "usual",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			PlanID:      1,
			PromocodeID: nil,
		}

		expectedErr := errors.New("storage error")
		mockSubRepo.EXPECT().StoreSub(gomock.Any(), gomock.Any()).Return(domain.SubID(0), expectedErr)

		subID, err := uc.NewSub(context.Background(), input)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if subID != 0 {
			t.Errorf("expected subID 0, got %d", subID)
		}
	})
}
