package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type SubRepo struct {
	p *pgxpool.Pool
}

func NewSubRepo(p *pgxpool.Pool) (*SubRepo, error) {
	if p == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &SubRepo{p: p}, nil
}

const (
	GetSubById = `
SELECT sub_id, service_id, price, start_date, end_date 
FROM subscriptions
WHERE sub_id = $1;
`
	GetUserBySubId = `
SELECT user_id FROM users_subs WHERE sub_id = $1;
`
	GetServiceNameById = `
SELECT service_name FROM services WHERE service_id = $1;	
`
	GetSubByUserId = `
SELECT sub_id FROM users_subs WHERE user_id = $1;	
`
	PutServiceName = `
INSERT INTO services (service_name) 
VALUES ($1)
ON CONFLICT (service_name) DO UPDATE SET service_name = EXCLUDED.service_name
RETURNING service_id;
`
	PutSub = `
INSERT INTO subscriptions
(service_id, price, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING sub_id;
`
	PutSubIdUserId = `
INSERT INTO users_subs
(sub_id, user_id)
VALUES ($1, $2);
`
	DeleteSub = `
DELETE FROM subscriptions WHERE sub_id = $1;
`
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
	tx, err := s.p.Begin(ctx)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var sub domain.Subscription
	var serviceId int
	if err := tx.QueryRow(ctx, GetSubById, int(subId)).Scan(
		&sub.SubId,
		&serviceId,
		&sub.Price,
		&sub.StartDate,
		&sub.EndDate,
	); err != nil {
		return domain.Subscription{}, err
	}
	if err := tx.QueryRow(ctx, GetUserBySubId, int(subId)).Scan(&sub.UserID); err != nil {
		return domain.Subscription{}, err
	}
	if err := tx.QueryRow(ctx, GetServiceNameById, serviceId).Scan(&sub.ServiceName); err != nil {
		return domain.Subscription{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sub, nil
}

func (s *SubRepo) UserSubs(ctx context.Context, userId uuid.UUID) ([]domain.Subscription, error) {
	res := make([]domain.Subscription, 0, 1)
	rows, err := s.p.Query(ctx, GetSubByUserId, userId)
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
	tx, err := s.p.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var serviceId int

	if err := tx.QueryRow(ctx, PutServiceName, sub.ServiceName).Scan(&serviceId); err != nil {
		return 0, fmt.Errorf("failed to get service_id: %w", err)
	}
	var subId int
	var enDateOrNil any = sub.EndDate
	if sub.EndDate.IsZero() {
		enDateOrNil = nil
	}
	if err := tx.QueryRow(ctx, PutSub, serviceId, sub.Price, sub.StartDate, enDateOrNil).Scan(&subId); err != nil {
		return 0, fmt.Errorf("failed to insert sub: %w", err)
	}
	_, err = tx.Exec(ctx, PutSubIdUserId, subId, sub.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user subscription: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.SubID(subId), nil
}

func (s *SubRepo) UpdateSub(ctx context.Context, sub domain.Subscription) error {
	tx, err := s.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	serviceId := -1
	if sub.ServiceName != "" {
		if err := tx.QueryRow(ctx, PutServiceName, sub.ServiceName).Scan(&serviceId); err != nil {
			return fmt.Errorf("failed to get service_id: %w", err)
		}
	}

	query := `UPDATE subscriptions SET`
	args := []any{}
	argPos := 1

	if serviceId != -1 {
		query += fmt.Sprintf(" service_id = $%d,", argPos)
		args = append(args, serviceId)
		argPos++
	}

	if sub.Price > 0 {
		query += fmt.Sprintf(" price = $%d,", argPos)
		args = append(args, sub.Price)
		argPos++
	}

	if !sub.StartDate.IsZero() {
		if sub.StartDate.Day() != 1 {
			return fmt.Errorf("bad data: day must be 1st (start)")
		}
		query += fmt.Sprintf(" start_date = $%d,", argPos)
		args = append(args, sub.StartDate)
		argPos++
	}

	if !sub.EndDate.IsZero() {
		if sub.EndDate.Day() != 1 {
			return fmt.Errorf("bad data: day must be 1st (end)")
		}
		query += fmt.Sprintf(" end_date = $%d,", argPos)
		args = append(args, sub.EndDate)
		argPos++
	}

	if argPos == 1 {
		return fmt.Errorf("no arguments to update")
	}
	query = strings.TrimSuffix(query, ",")

	query += fmt.Sprintf(" WHERE sub_id = $%d", argPos)
	args = append(args, int(sub.SubId))

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("fail: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("can't finish transaction: %w", err)
	}
	return nil
}

func (s *SubRepo) DeleteSub(ctx context.Context, subId domain.SubID) error {
	_, err := s.p.Exec(ctx, DeleteSub, int(subId))
	return err
}

func (s *SubRepo) SubsTotalCosts(ctx context.Context, filter domain.SubsFilter) (int, []domain.SubID, error) {
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
	subIds := make([]domain.SubID, 0, len(allSubs))

	for _, sub := range allSubs {
		subIds = append(subIds, sub.SubId)

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
			return 0, nil, fmt.Errorf("invalid dates")
		}
		sumCost += sub.Price * months
	}
	return sumCost, subIds, nil

}
