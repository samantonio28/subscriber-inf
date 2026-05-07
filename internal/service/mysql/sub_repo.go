package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
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
	getSubById = `
SELECT s.sub_id, s.user_id, s.plan_id, s.promocode_id, s.price, s.sub_type, s.start_date, s.end_date,
       sv.service_name
FROM subscriptions s
LEFT JOIN subscription_plans sp ON s.plan_id = sp.plan_id
LEFT JOIN services sv ON sp.service_id = sv.service_id
WHERE s.sub_id = ?;
`
	getServiceNameById = `
SELECT service_name FROM services WHERE service_id = ?;
`
	getSubByUserId = `
SELECT sub_id FROM subscriptions WHERE user_id = ?;
`
	putServiceName = `
INSERT INTO services (service_name)
VALUES (?)
ON DUPLICATE KEY UPDATE service_name = VALUES(service_name);
`
	getServiceIdByPlanId = `
SELECT service_id FROM subscription_plans WHERE plan_id = ?;
`
	putSub = `
INSERT INTO subscriptions
(user_id, service_id, plan_id, promocode_id, price, sub_type, start_date, end_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`
	deleteSub = `
DELETE FROM subscriptions WHERE sub_id = ?;
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
	if err := tx.QueryRowContext(ctx, getSubById, int(subId)).Scan(
		&sub.SubId,
		&sub.UserID,
		&sub.PlanID,
		&promocodeID,
		&sub.Price,
		&subType,
		&sub.StartDate,
		&enDate,
		&sub.ServiceName,
	); err != nil {
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
	rows, err := s.db.QueryContext(ctx, getSubByUserId, userId)
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

	// Получить service_id по plan_id
	var serviceID int
	err = tx.QueryRowContext(ctx, getServiceIdByPlanId, sub.PlanID).Scan(&serviceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get service_id for plan %d: %w", sub.PlanID, err)
	}

	var enDateOrNil any = sub.EndDate
	if sub.EndDate.IsZero() {
		enDateOrNil = nil
	}
	var promoIDOrNil any = sub.PromocodeID
	if sub.PromocodeID == nil {
		promoIDOrNil = nil
	}
	result, err := tx.ExecContext(ctx, putSub, sub.UserID, serviceID, sub.PlanID, promoIDOrNil, sub.Price, sub.SubType.String(), sub.StartDate, enDateOrNil)
	if err != nil {
		return 0, fmt.Errorf("failed to insert sub: %w", err)
	}
	subId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.SubID(subId), nil
}

func (s *SubRepo) UpdateSub(ctx context.Context, sub domain.Subscription) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	subToCheck, err := s.Sub(ctx, sub.SubId)
	if err != nil {
		return fmt.Errorf("sub does not exist: %w", err)
	}
	if sub.UserID == uuid.Nil {
		return fmt.Errorf("bad data: user_id cannot be nil")
	}
	if sub.UserID != subToCheck.UserID {
		return fmt.Errorf("bad data: user_id cannot be changed")
	}

	query := `UPDATE subscriptions SET`
	args := []any{}
	argPos := 1

	if sub.PlanID != 0 {
		query += " plan_id = ?,"
		args = append(args, sub.PlanID)
		argPos++
	}

	if sub.PromocodeID != nil {
		query += " promocode_id = ?,"
		args = append(args, *sub.PromocodeID)
		argPos++
	}

	if sub.Price > 0 {
		query += " price = ?,"
		args = append(args, sub.Price)
		argPos++
	}

	if subtype, err := domain.NewSubscriptionType(sub.SubType.String()); err == nil {
		query += " sub_type = ?,"
		args = append(args, subtype.String())
		argPos++
	}

	if !sub.StartDate.IsZero() {
		if sub.StartDate.Day() != 1 {
			return fmt.Errorf("bad data: day must be 1st (start)")
		}
		query += " start_date = ?,"
		args = append(args, sub.StartDate)
		argPos++
	}

	if !sub.EndDate.IsZero() {
		if sub.EndDate.Day() != 1 {
			return fmt.Errorf("bad data: day must be 1st (end)")
		}
		if !sub.StartDate.IsZero() && sub.EndDate.Before(sub.StartDate) ||
			sub.StartDate.IsZero() && sub.EndDate.Before(subToCheck.StartDate) {
			return fmt.Errorf("bad data: end date must be after start date")
		}
		query += " end_date = ?,"
		args = append(args, sub.EndDate)
		argPos++
	}

	if argPos == 1 {
		return fmt.Errorf("no arguments to update")
	}
	query = strings.TrimSuffix(query, ",")

	query += fmt.Sprintf(" WHERE sub_id = ?")
	args = append(args, int(sub.SubId))

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("fail: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("can't finish transaction: %w", err)
	}
	return nil
}

func (s *SubRepo) DeleteSub(ctx context.Context, subId domain.SubID) error {
	res, err := s.db.ExecContext(ctx, deleteSub, int(subId))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNoSubscriptionDeleted
	}
	return nil
}

func (s *SubRepo) SubsTotalCosts(ctx context.Context, filter domain.SubsFilter) (int, []domain.Subscription, error) {
	solve := "backend"
	switch solve {
	case "db":
		return s.subsTotalCostsDB(ctx, filter)
	case "backend":
		return s.subsTotalCostsBackend(ctx, filter)
	default:
		return 0, nil, nil
	}
}

func (s *SubRepo) subsTotalCostsBackend(ctx context.Context, filter domain.SubsFilter) (int, []domain.Subscription, error) {
	if filter.UserID == uuid.Nil || filter.StartDate.IsZero() || !filter.EndDate.IsZero() && filter.EndDate.Before(filter.StartDate) {
		return 0, nil, fmt.Errorf("user id and start date is required || end date must be after start date")
	}

	allSubs, err := s.UserSubs(ctx, filter.UserID)
	if err != nil {
		return 0, nil, fmt.Errorf("can't get user subs: %w", err)
	}

	if filter.EndDate.IsZero() {
		filter.EndDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	sumCost := 0
	subs := make([]domain.Subscription, 0, len(allSubs))

	for _, sub := range allSubs {
		if filter.SubType.String() != "" && sub.SubType != filter.SubType {
			continue
		}
		if sub.ServiceName != filter.ServiceName {
			continue
		}
		st := sub.StartDate
		en := sub.EndDate
		if st.Before(filter.StartDate) {
			st = filter.StartDate
		}
		if en.IsZero() || filter.EndDate.Before(en) {
			en = filter.EndDate
		}

		months := utils.MonthToInt(en.Month()) - utils.MonthToInt(st.Month()) + 12*(en.Year()-st.Year())
		if months < 0 {
			continue
		}
		sumCost += sub.Price * months
		subs = append(subs, sub)
	}
	return sumCost, subs, nil
}

func (s *SubRepo) subsTotalCostsDB(_ context.Context, _ domain.SubsFilter) (int, []domain.Subscription, error) {
	return 0, nil, nil
}