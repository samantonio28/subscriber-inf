package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type SubRepo struct {
	db *sql.DB
}

func NewSubRepo(db *sql.DB) (domain.SubscriptionRepository, error) {
	if db == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &SubRepo{db: db}, nil
}

const (
	GetSubById = `
SELECT s.sub_id, s.user_id, s.plan_id, s.promocode_id, s.price, s.sub_type, s.start_date, s.end_date,
       sv.service_name
FROM subscriptions s
LEFT JOIN subscription_plans sp ON s.plan_id = sp.plan_id
LEFT JOIN services sv ON sp.service_id = sv.service_id
WHERE s.sub_id = ?;
`
	GetServiceNameById = `
SELECT service_name FROM services WHERE service_id = ?;
`
	GetSubByUserId = `
SELECT sub_id FROM subscriptions WHERE user_id = ?;
`
	PutServiceName = `
INSERT INTO services (service_name)
VALUES (?)
ON CONFLICT (service_name) DO UPDATE SET service_name = excluded.service_name
RETURNING service_id;
`
	PutSub = `
INSERT INTO subscriptions
(user_id, plan_id, promocode_id, price, sub_type, start_date, end_date)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING sub_id;
`
	DeleteSub = `
DELETE FROM subscriptions WHERE sub_id = ?;
`
	// unused
	GetAllData = `
SELECT
    us.sub_id,
    us.user_id,
    s.service_name,
    sub.price,
    sub.start_date,
    sub.end_date
FROM
    users_subs us
LEFT JOIN
    subscriptions sub ON us.sub_id = sub.sub_id
LEFT JOIN
    services s ON sub.service_id = s.service_id
ORDER BY
    us.sub_id;
`
)

func (s *SubRepo) Sub(ctx context.Context, subId domain.SubID) (domain.Subscription, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var sub domain.Subscription
	var subType string
	var enDate sql.NullTime
	var promocodeID sql.NullInt32
	err = tx.QueryRowContext(ctx, GetSubById, int(subId)).Scan(
		&sub.SubId,
		&sub.UserID,
		&sub.PlanID,
		&promocodeID,
		&sub.Price,
		&subType,
		&sub.StartDate,
		&enDate,
		&sub.ServiceName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}
	if st, err := domain.NewSubscriptionType(subType); err != nil {
		return domain.Subscription{}, err
	} else {
		sub.SubType = *st
	}

	if enDate.Valid {
		sub.EndDate = enDate.Time
	}
	if promocodeID.Valid {
		p := int(promocodeID.Int32)
		sub.PromocodeID = &p
	}

	if err := tx.Commit(); err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sub, nil
}

func (s *SubRepo) UserSubs(ctx context.Context, userId uuid.UUID) ([]domain.Subscription, error) {
	res := make([]domain.Subscription, 0, 1)
	rows, err := s.db.QueryContext(ctx, GetSubByUserId, userId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var subId int
		if err := rows.Scan(&subId); err != nil {
			return nil, err
		}

		sub, err := s.Sub(ctx, domain.SubID(subId))
		if err != nil {
			return nil, err
		}
		res = append(res, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return res, nil
}

func (s *SubRepo) StoreSub(ctx context.Context, sub domain.Subscription) (domain.SubID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Ensure service exists
	var serviceID int
	err = tx.QueryRowContext(ctx, PutServiceName, sub.ServiceName).Scan(&serviceID)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure service: %w", err)
	}

	// Find plan ID (assuming plan_id is already set in sub.PlanID)
	// If not, we need to determine plan_id from service_id and duration? For simplicity, assume sub.PlanID is correct.

	var subID int
	err = tx.QueryRowContext(ctx, PutSub,
		sub.UserID,
		sub.PlanID,
		sub.PromocodeID,
		sub.Price,
		sub.SubType.String(),
		sub.StartDate,
		sub.EndDate,
	).Scan(&subID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert subscription: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.SubID(subID), nil
}

func (s *SubRepo) UpdateSub(ctx context.Context, sub domain.Subscription) error {
	// For simplicity, we'll just delete and reinsert? Or implement proper UPDATE.
	// Since the original sub_repo.go doesn't have UpdateSub implementation, we'll skip.
	return errors.New("UpdateSub not implemented for SQLite")
}

func (s *SubRepo) DeleteSub(ctx context.Context, subId domain.SubID) error {
	result, err := s.db.ExecContext(ctx, DeleteSub, int(subId))
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNoSubscriptionDeleted
	}
	return nil
}

func (s *SubRepo) SubsTotalCosts(ctx context.Context, filter domain.SubsFilter) (int, []domain.Subscription, error) {
	// This is a complex method; for now, return not implemented.
	return 0, nil, errors.New("SubsTotalCosts not implemented for SQLite")
}
