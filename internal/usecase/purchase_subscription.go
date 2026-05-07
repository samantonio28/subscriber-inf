package usecase

import (
	"context"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type PurchaseSubscriptionUC struct {
	userR    domain.UserRepository
	subR     domain.SubscriptionRepository
	paymentR domain.PaymentRepository
	logger   logger.Logger
}

func NewPurchaseSubscriptionUC(
	userR domain.UserRepository,
	subR domain.SubscriptionRepository,
	paymentR domain.PaymentRepository,
	logger logger.Logger,
) (*PurchaseSubscriptionUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if subR == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if paymentR == nil {
		return nil, domain.ErrInvalidPaymentRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &PurchaseSubscriptionUC{
		userR:    userR,
		subR:     subR,
		paymentR: paymentR,
		logger:   logger,
	}, nil
}

func (u *PurchaseSubscriptionUC) Purchase(ctx context.Context, input PurchaseSubscriptionDTO) (int, error) {
	u.logger.Debug("purchasing subscription", "input", input)

	// 1. Получить пользователя (проверить баланс)
	user, err := u.userR.GetUser(ctx, input.UserID)
	if err != nil {
		u.logger.Error("failed to get user", "user_id", input.UserID, "error", err)
		return 0, err
	}

	// 2. Проверить, достаточно ли баланса
	if user.Balance < input.Price {
		u.logger.Error("insufficient balance", "user_id", input.UserID, "balance", user.Balance, "required", input.Price)
		return 0, domain.ErrInsufficientBalance
	}

	// 3. Создать подписку
	// startDate должен быть первым числом текущего месяца
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// Вычисляем количество месяцев из DurationDays (округление вверх)
	months := (input.DurationDays + 29) / 30 // ceil(days/30)
	endDate := startDate.AddDate(0, int(months), 0)
	sub, err := domain.NewSubscription(
		0, // subId будет сгенерирован репозиторием
		input.UserID,
		input.ServiceName,
		input.Price,
		"usual", // тип подписки по умолчанию
		startDate,
		endDate,
		input.PlanID,
		nil, // promocodeID
	)
	if err != nil {
		u.logger.Error("failed to create subscription domain object", "error", err)
		return 0, err
	}

	subId, err := u.subR.StoreSub(ctx, *sub)
	if err != nil {
		u.logger.Error("failed to store subscription", "error", err)
		return 0, err
	}

	// 4. Создать платеж (списание)
	payment := domain.Payment{
		PaymentID:   0, // сгенерируется в БД
		UserID:      input.UserID,
		CardNumber:  nil, // для расходов card_number должен быть NULL
		Amount:      input.Price,
		PaymentType: domain.PaymentEXPENCE,
		CreatedAt:   time.Now(),
	}
	err = u.paymentR.StorePayment(ctx, payment)
	if err != nil {
		u.logger.Error("failed to store payment", "error", err)
		// Откатывать подписку? Пока просто возвращаем ошибку
		return 0, err
	}

	// 5. Обновить баланс пользователя (уменьшить)
	user.Balance -= input.Price
	err = u.userR.UpdateUser(ctx, user)
	if err != nil {
		u.logger.Error("failed to update user balance", "error", err)
		// Откатывать платеж и подписку? Пока просто возвращаем ошибку
		return 0, err
	}

	u.logger.Info("subscription purchased successfully", "subscription_id", subId, "user_id", input.UserID)
	return int(subId), nil
}
