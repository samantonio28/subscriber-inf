package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
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
)

func (s *SubRepo) Sub(ctx context.Context, subId domain.SubID) (*domain.Subscription, error) {
	var sub domain.Subscription
	var serviceId int
	if err := s.p.QueryRow(ctx, GetSubById, int(subId)).Scan(
		&sub.SubId,
		&serviceId,
		&sub.Price,
		&sub.StartDate,
		&sub.EndDate,
	); err != nil {
		return nil, err
	}
	if err := s.p.QueryRow(ctx, GetUserBySubId, int(subId)).Scan(&sub.UserID); err != nil {
		return nil, err
	}
	if err := s.p.QueryRow(ctx, GetServiceNameById, serviceId).Scan(&sub.ServiceName); err != nil {
		return nil, err
	}
	return &sub, nil
}

// func (s *SubRepo) UserSubs(ctx context.Context, userId uuid.UUID) ([]*domain.Subscription, error) {

// }

// func (s *SubRepo) StoreSub(ctx context.Context, sub domain.Subscription) (domain.SubID, error) {

// }

// func (s *SubRepo) UpdateSub(ctx context.Context, sub domain.Subscription) error {

// }

// func (s *SubRepo) DeleteSub(ctx context.Context, subId domain.SubID) error {

// }

// func (s *SubRepo) SubsTotalCosts(ctx context.Context, filter domain.SubsFilter) (int, []domain.SubID, error) {

// }
