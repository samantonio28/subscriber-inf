package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type UserServiceRepo struct {
	p *pgxpool.Pool
}

func NewUserServiceRepo(p *pgxpool.Pool) (domain.UserServiceRepository, error) {
	if p == nil {
		return nil, fmt.Errorf("pgxpool.Pool is nil")
	}
	return &UserServiceRepo{p: p}, nil
}

const (
	addUserService = `
INSERT INTO user_services (user_id, service_id)
VALUES ($1, $2)
ON CONFLICT (user_id, service_id) DO NOTHING
`
	removeUserService = `
DELETE FROM user_services
WHERE user_id = $1 AND service_id = $2
`
	listByUser = `
SELECT service_id
FROM user_services
WHERE user_id = $1
ORDER BY created_at DESC
`
	listByService = `
SELECT user_id
FROM user_services
WHERE service_id = $1
ORDER BY created_at DESC
`
	existsUserService = `
SELECT EXISTS(
	SELECT 1 FROM user_services
	WHERE user_id = $1 AND service_id = $2
)
`
)

func (r *UserServiceRepo) Add(ctx context.Context, userID uuid.UUID, serviceID int) error {
	_, err := r.p.Exec(ctx, addUserService, userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to add user service: %w", err)
	}
	return nil
}

func (r *UserServiceRepo) Remove(ctx context.Context, userID uuid.UUID, serviceID int) error {
	_, err := r.p.Exec(ctx, removeUserService, userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to remove user service: %w", err)
	}
	return nil
}

func (r *UserServiceRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]int, error) {
	rows, err := r.p.Query(ctx, listByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user services: %w", err)
	}
	defer rows.Close()

	serviceIDs := make([]int, 0)
	for rows.Next() {
		var serviceID int
		if err := rows.Scan(&serviceID); err != nil {
			return nil, fmt.Errorf("failed to scan service_id: %w", err)
		}
		serviceIDs = append(serviceIDs, serviceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return serviceIDs, nil
}

func (r *UserServiceRepo) ListByService(ctx context.Context, serviceID int) ([]uuid.UUID, error) {
	rows, err := r.p.Query(ctx, listByService, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query service users: %w", err)
	}
	defer rows.Close()

	userIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user_id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return userIDs, nil
}

func (r *UserServiceRepo) Exists(ctx context.Context, userID uuid.UUID, serviceID int) (bool, error) {
	var exists bool
	err := r.p.QueryRow(ctx, existsUserService, userID, serviceID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}
