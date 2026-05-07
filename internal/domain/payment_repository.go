package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentType string

const (
	PaymentINCOME  PaymentType = "income"
	PaymentEXPENCE PaymentType = "expence"
)

type Payment struct {
	PaymentID   int
	UserID      uuid.UUID
	CardNumber  *string
	Amount      int
	PaymentType PaymentType
	CreatedAt   time.Time
}

type PaymentRepository interface {
	StorePayment(ctx context.Context, payment Payment) error
	GetUserPayments(ctx context.Context, userID uuid.UUID) ([]Payment, error)
}
