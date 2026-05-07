package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) (domain.PaymentRepository, error) {
	if db == nil {
		return nil, domain.ErrInvalidPaymentRepo
	}
	return &PaymentRepo{db: db}, nil
}

const (
	storePaymentQuery = `
INSERT INTO payments (user_id, card_number, amount, paym_type)
VALUES (?, ?, ?, ?)
`
	getUserPaymentsQuery = `
SELECT paym_id, user_id, card_number, amount, paym_type
FROM payments
WHERE user_id = ?
ORDER BY paym_id;
`
)

func (r *PaymentRepo) StorePayment(ctx context.Context, payment domain.Payment) error {
	result, err := r.db.ExecContext(ctx, storePaymentQuery,
		payment.UserID,
		payment.CardNumber,
		payment.Amount,
		payment.PaymentType,
	)
	if err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}
	_, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	return nil
}

func (r *PaymentRepo) GetUserPayments(ctx context.Context, userID uuid.UUID) ([]domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx, getUserPaymentsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var paymType string
		var cardNumber sql.NullString
		err := rows.Scan(&p.PaymentID, &p.UserID, &cardNumber, &p.Amount, &paymType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row: %w", err)
		}
		if cardNumber.Valid {
			p.CardNumber = &cardNumber.String
		} else {
			p.CardNumber = nil
		}
		p.PaymentType = domain.PaymentType(paymType)
		p.CreatedAt = time.Time{} // not stored
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment rows: %w", err)
	}
	return payments, nil
}