package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type SubscriptionPlanRepo struct {
	db *sql.DB
}

func NewSubscriptionPlanRepo(db *sql.DB) (domain.SubscriptionPlanRepository, error) {
	if db == nil {
		return nil, domain.ErrInvalidSubPlanRepo
	}
	return &SubscriptionPlanRepo{db: db}, nil
}

const (
	getPlanByID = `
SELECT plan_id, service_id, name, duration_days, price
FROM subscription_plans
WHERE plan_id = ?;
`
	getPlansByService = `
SELECT plan_id, service_id, name, duration_days, price
FROM subscription_plans
WHERE service_id = ?;
`
	insertPlan = `
INSERT INTO subscription_plans (service_id, name, duration_days, price)
VALUES (?, ?, ?, ?)
`
	updatePlan = `
UPDATE subscription_plans
SET service_id = ?, name = ?, duration_days = ?, price = ?
WHERE plan_id = ?;
`
	deletePlan = `
DELETE FROM subscription_plans WHERE plan_id = ?;
`
	getAllPlans = `
SELECT plan_id, service_id, name, duration_days, price
FROM subscription_plans;
`
)

func (r *SubscriptionPlanRepo) GetByID(ctx context.Context, id domain.PlanID) (domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.QueryRowContext(ctx, getPlanByID, int(id)).Scan(
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
	rows, err := r.db.QueryContext(ctx, getPlansByService, serviceID)
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

func (r *SubscriptionPlanRepo) GetAll(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	rows, err := r.db.QueryContext(ctx, getAllPlans)
	if err != nil {
		return nil, fmt.Errorf("failed to query all plans: %w", err)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	result, err := tx.ExecContext(ctx, insertPlan,
		plan.ServiceID,
		plan.Name,
		plan.DurationDays,
		plan.Price,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert plan: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.PlanID(id), nil
}

func (r *SubscriptionPlanRepo) Update(ctx context.Context, plan domain.SubscriptionPlan) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	_, err = tx.ExecContext(ctx, updatePlan,
		plan.ServiceID,
		plan.Name,
		plan.DurationDays,
		plan.Price,
		int(plan.PlanID),
	)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *SubscriptionPlanRepo) Delete(ctx context.Context, id domain.PlanID) error {
	res, err := r.db.ExecContext(ctx, deletePlan, int(id))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNoPlansDeleted
	}
	return nil
}