package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// Лондонский стиль: все три репозитория заменены моками.
func TestNewPurchaseSubscriptionUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)
	mockSub := mock.NewMockSubscriptionRepository(ctrl)
	mockPayment := mock.NewMockPaymentRepository(ctrl)
	nop := &testutil.NopLogger{}

	t.Run("successful creation", func(t *testing.T) {
		uc, err := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, nop)
		if err != nil || uc == nil {
			t.Fatalf("expected ok, got uc=%v err=%v", uc, err)
		}
	})
	t.Run("nil user repository", func(t *testing.T) {
		_, err := NewPurchaseSubscriptionUC(nil, mockSub, mockPayment, nop)
		if !errors.Is(err, domain.ErrInvalidUserRepo) {
			t.Fatalf("expected ErrInvalidUserRepo, got %v", err)
		}
	})
	t.Run("nil subscription repository", func(t *testing.T) {
		_, err := NewPurchaseSubscriptionUC(mockUser, nil, mockPayment, nop)
		if !errors.Is(err, domain.ErrInvalidSubRepo) {
			t.Fatalf("expected ErrInvalidSubRepo, got %v", err)
		}
	})
	t.Run("nil payment repository", func(t *testing.T) {
		_, err := NewPurchaseSubscriptionUC(mockUser, mockSub, nil, nop)
		if !errors.Is(err, domain.ErrInvalidPaymentRepo) {
			t.Fatalf("expected ErrInvalidPaymentRepo, got %v", err)
		}
	})
	t.Run("nil logger", func(t *testing.T) {
		_, err := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, nil)
		if !errors.Is(err, domain.ErrInvalidLogger) {
			t.Fatalf("expected ErrInvalidLogger, got %v", err)
		}
	})
}

func TestPurchaseSubscriptionUC_Purchase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("successful purchase stores payment and decreases balance", func(t *testing.T) {
		// Arrange
		user := testutil.TestUser(1000)
		input := PurchaseSubscriptionDTO{
			UserID:       user.UserID,
			ServiceName:  "Netflix",
			PlanID:       1,
			Price:        300,
			DurationDays: 30,
		}

		mockUser := mock.NewMockUserRepository(ctrl)
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPayment := mock.NewMockPaymentRepository(ctrl)

		mockUser.EXPECT().GetUser(ctx, user.UserID).Return(user, nil)
		mockSub.EXPECT().StoreSub(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, s domain.Subscription) (domain.SubID, error) {
				if s.Price != 300 || s.ServiceName != "Netflix" {
					t.Errorf("unexpected stored subscription: %+v", s)
				}
				return domain.SubID(42), nil
			},
		)
		mockPayment.EXPECT().StorePayment(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, p domain.Payment) error {
				if p.Amount != 300 || p.PaymentType != domain.PaymentEXPENCE {
					t.Errorf("unexpected stored payment: %+v", p)
				}
				return nil
			},
		)
		mockUser.EXPECT().UpdateUser(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u domain.User) error {
				if u.Balance != 700 {
					t.Errorf("expected balance 700, got %d", u.Balance)
				}
				return nil
			},
		)

		uc, err := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, &testutil.NopLogger{})
		if err != nil {
			t.Fatalf("failed to create usecase: %v", err)
		}

		// Act
		id, err := uc.Purchase(ctx, input)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if id != 42 {
			t.Errorf("expected sub id 42, got %d", id)
		}
	})

	t.Run("insufficient balance rejected", func(t *testing.T) {
		user := testutil.TestUser(100)
		input := PurchaseSubscriptionDTO{UserID: user.UserID, ServiceName: "S", PlanID: 1, Price: 300, DurationDays: 30}

		mockUser := mock.NewMockUserRepository(ctrl)
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPayment := mock.NewMockPaymentRepository(ctrl)

		mockUser.EXPECT().GetUser(ctx, user.UserID).Return(user, nil)

		uc, _ := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, &testutil.NopLogger{})

		id, err := uc.Purchase(ctx, input)
		if !errors.Is(err, domain.ErrInsufficientBalance) || id != 0 {
			t.Fatalf("expected ErrInsufficientBalance with id 0, got err=%v id=%d", err, id)
		}
	})

	t.Run("store payment failure propagated", func(t *testing.T) {
		user := testutil.TestUser(1000)
		input := PurchaseSubscriptionDTO{UserID: user.UserID, ServiceName: "S", PlanID: 1, Price: 100, DurationDays: 30}

		mockUser := mock.NewMockUserRepository(ctrl)
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPayment := mock.NewMockPaymentRepository(ctrl)

		mockUser.EXPECT().GetUser(ctx, user.UserID).Return(user, nil)
		mockSub.EXPECT().StoreSub(ctx, gomock.Any()).Return(domain.SubID(1), nil)
		mockPayment.EXPECT().StorePayment(ctx, gomock.Any()).Return(errors.New("payment failed"))

		uc, _ := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, &testutil.NopLogger{})

		_, err := uc.Purchase(ctx, input)
		if err == nil || err.Error() != "payment failed" {
			t.Fatalf("expected payment failed, got %v", err)
		}
	})
}
