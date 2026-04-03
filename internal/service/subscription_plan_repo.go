package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type SubscriptionPlanRepo struct {
	p *pgxpool.Pool
}

func NewSubscriptionPlanRepo(p *pgxpool.Pool) (domain.SubscriptionPlanRepository, error) {
	if p == nil {
		return nil, domain.ErrInvalidSubPlanRepo // reuse error
	}
	return &SubscriptionPlanRepo{p: p}, nil
}

const (
	getPlanByID = `
SELECT plan_id, service_id, name, duration_days, price
FROM subscription_plans
WHERE plan_id = $1;
`
	getPlansByService = `
SELECT plan_id, service_id, name, duration_days, price
FROM subscription_plans
WHERE service_id = $1;
`
	insertPlan = `
INSERT INTO subscription_plans (service_id, name, duration_days, price)
VALUES ($1, $2, $3, $4)
RETURNING plan_id;
`
	updatePlan = `
UPDATE subscription_plans
SET service_id = $2, name = $3, duration_days = $4, price = $5
WHERE plan_id = $1;
`
	deletePlan = `
DELETE FROM subscription_plans WHERE plan_id = $1;
`
)

func (r *SubscriptionPlanRepo) GetByID(ctx context.Context, id domain.PlanID) (domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.p.QueryRow(ctx, getPlanByID, int(id)).Scan(
		&plan.PlanID,
		&plan.ServiceID,
		&plan.Name,
		&plan.DurationDays,
		&plan.Price,
	)
	if err != nil {
		return domain.SubscriptionPlan{}, fmt.Errorf("failed to query plan: %w", err)
	}
	return plan, nil
}

func (r *SubscriptionPlanRepo) GetByService(ctx context.Context, serviceID int) ([]domain.SubscriptionPlan, error) {
	rows, err := r.p.Query(ctx, getPlansByService, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []domain.SubscriptionPlan
	for rows.Next() {
		var plan domain.SubscriptionPlan
		err := rows.Scan(
			&plan.PlanID,
			&plan.ServiceID,
			&plan.Name,
			&plan.DurationDays,
			&plan.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return plans, nil
}

func (r *SubscriptionPlanRepo) Create(ctx context.Context, plan domain.SubscriptionPlan) (domain.PlanID, error) {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var id int
	err = tx.QueryRow(ctx, insertPlan,
		plan.ServiceID,
		plan.Name,
		plan.DurationDays,
		plan.Price,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert plan: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.PlanID(id), nil
}

func (r *SubscriptionPlanRepo) Update(ctx context.Context, plan domain.SubscriptionPlan) error {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	_, err = tx.Exec(ctx, updatePlan,
		int(plan.PlanID),
		plan.ServiceID,
		plan.Name,
		plan.DurationDays,
		plan.Price,
	)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *SubscriptionPlanRepo) Delete(ctx context.Context, id domain.PlanID) error {
	res, err := r.p.Exec(ctx, deletePlan, int(id))
	if err != nil {
		return err
	}
	if r := res.RowsAffected(); r == 0 {
		return domain.ErrNoPlansDeleted
	}
	return nil
}
