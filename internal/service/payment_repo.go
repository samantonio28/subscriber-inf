package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type PaymentRepo struct {
	p *pgxpool.Pool
}

func NewPaymentRepo(p *pgxpool.Pool) (domain.PaymentRepository, error) {
	if p == nil {
		return nil, domain.ErrInvalidPaymentRepo
	}
	return &PaymentRepo{p: p}, nil
}

const (
	storePaymentQuery = `
INSERT INTO payments (user_id, card_number, amount, paym_type)
VALUES ($1, $2, $3, $4)
RETURNING paym_id;
`
	getUserPaymentsQuery = `
SELECT paym_id, user_id, card_number, amount, paym_type
FROM payments
WHERE user_id = $1
ORDER BY paym_id;
`
)

func (r *PaymentRepo) StorePayment(ctx context.Context, payment domain.Payment) error {
	var paymentID int
	err := r.p.QueryRow(ctx, storePaymentQuery,
		payment.UserID,
		payment.CardNumber,
		payment.Amount,
		payment.PaymentType,
	).Scan(&paymentID)
	if err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}
	// payment.PaymentID already set? we can assign if needed
	return nil
}

func (r *PaymentRepo) GetUserPayments(ctx context.Context, userID uuid.UUID) ([]domain.Payment, error) {
	rows, err := r.p.Query(ctx, getUserPaymentsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var paymType string
		var cardNumber *string
		err := rows.Scan(&p.PaymentID, &p.UserID, &cardNumber, &p.Amount, &paymType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row: %w", err)
		}
		p.CardNumber = cardNumber
		p.PaymentType = domain.PaymentType(paymType)
		p.CreatedAt = time.Time{} // not stored
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment rows: %w", err)
	}
	return payments, nil
}
