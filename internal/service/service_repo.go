package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type ServiceRepo struct {
	p *pgxpool.Pool
}

func NewServiceRepo(p *pgxpool.Pool) (domain.ServiceRepository, error) {
	if p == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &ServiceRepo{p: p}, nil
}

const serviceColumns = `service_id, service_name, COALESCE(users_count, 0), COALESCE(has_promocodes, false)`

func (r *ServiceRepo) StoreService(ctx context.Context, s domain.Service) (int, error) {
	var id int
	err := r.p.QueryRow(ctx, `
		INSERT INTO services (service_name, users_count, has_promocodes)
		VALUES ($1, $2, $3)
		RETURNING service_id`,
		s.ServiceName, s.UsersCount, s.HasPromocodes,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ServiceRepo) scanService(row pgx.Row) (domain.Service, error) {
	var s domain.Service
	err := row.Scan(&s.ServiceID, &s.ServiceName, &s.UsersCount, &s.HasPromocodes)
	return s, err
}

func (r *ServiceRepo) GetService(ctx context.Context, serviceID int) (domain.Service, error) {
	return r.scanService(r.p.QueryRow(ctx, `SELECT `+serviceColumns+` FROM services WHERE service_id = $1`, serviceID))
}

func (r *ServiceRepo) GetServiceByName(ctx context.Context, name string) (domain.Service, error) {
	return r.scanService(r.p.QueryRow(ctx, `SELECT `+serviceColumns+` FROM services WHERE service_name = $1`, name))
}

func (r *ServiceRepo) ListServices(ctx context.Context) ([]domain.Service, error) {
	rows, err := r.p.Query(ctx, `SELECT `+serviceColumns+` FROM services ORDER BY service_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Service, 0)
	for rows.Next() {
		s, err := r.scanService(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *ServiceRepo) UpdateService(ctx context.Context, s domain.Service) error {
	tag, err := r.p.Exec(ctx, `
		UPDATE services SET service_name = $1, users_count = $2, has_promocodes = $3
		WHERE service_id = $4`,
		s.ServiceName, s.UsersCount, s.HasPromocodes, s.ServiceID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("service not found")
	}
	return nil
}

func (r *ServiceRepo) DeleteService(ctx context.Context, serviceID int) error {
	tag, err := r.p.Exec(ctx, `DELETE FROM services WHERE service_id = $1`, serviceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("service not found")
	}
	return nil
}
