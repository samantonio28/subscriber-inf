package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type PromocodeRepo struct {
	p *pgxpool.Pool
}

func NewPromocodeRepo(p *pgxpool.Pool) (domain.PromocodeRepository, error) {
	if p == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	return &PromocodeRepo{p: p}, nil
}

const (
	getPromocodeByID = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE promocode_id = $1;
`
	getPromocodeByCode = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE promocode = $1;
`
	getPromocodesByService = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE service_id = $1;
`
	insertPromocode = `
INSERT INTO promocodes (service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING promocode_id;
`
	updatePromocode = `
UPDATE promocodes
SET service_id = $2, promocode = $3, plan_id = $4, sub_id = $5, expires_at = $6, created_at = $7, discount = $8, max_uses = $9, cur_uses = $10, status = $11, duration_days = $12
WHERE promocode_id = $1;
`
	deletePromocode = `
DELETE FROM promocodes WHERE promocode_id = $1;
`
	incrementUses = `
UPDATE promocodes
SET cur_uses = cur_uses + 1
WHERE promocode_id = $1;
`
	getAllPromocodes = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes;
`
)

func (r *PromocodeRepo) GetByID(ctx context.Context, id domain.PromocodeID) (domain.Promocode, error) {
	var pc domain.Promocode
	var planID pgtype.Int4
	var subID pgtype.Int4
	var expiresAt pgtype.Date
	var createdAt pgtype.Timestamp
	var status string

	err := r.p.QueryRow(ctx, getPromocodeByID, int(id)).Scan(
		&pc.PromocodeID,
		&pc.ServiceID,
		&pc.Value,
		&planID,
		&subID,
		&expiresAt,
		&createdAt,
		&pc.Discount,
		&pc.MaxUses,
		&pc.CurUses,
		&status,
		&pc.DurationDays,
	)
	if err != nil {
		return domain.Promocode{}, fmt.Errorf("failed to query promocode: %w", err)
	}

	if planID.Valid {
		val := int(planID.Int32)
		pc.PlanID = &val
	}
	if subID.Valid {
		val := int(subID.Int32)
		pc.SubID = &val
	}
	if expiresAt.Valid {
		pc.ExpiresAt = expiresAt.Time
	}
	if createdAt.Valid {
		pc.CreatedAt = createdAt.Time
	}

	statusObj, err := domain.NewPromocodeStatus(status)
	if err != nil {
		return domain.Promocode{}, err
	}
	pc.Status = statusObj

	return pc, nil
}

func (r *PromocodeRepo) GetByCode(ctx context.Context, code string) (domain.Promocode, error) {
	var pc domain.Promocode
	var planID pgtype.Int4
	var subID pgtype.Int4
	var expiresAt pgtype.Date
	var createdAt pgtype.Timestamp
	var status string

	err := r.p.QueryRow(ctx, getPromocodeByCode, code).Scan(
		&pc.PromocodeID,
		&pc.ServiceID,
		&pc.Value,
		&planID,
		&subID,
		&expiresAt,
		&createdAt,
		&pc.Discount,
		&pc.MaxUses,
		&pc.CurUses,
		&status,
		&pc.DurationDays,
	)
	if err != nil {
		return domain.Promocode{}, fmt.Errorf("failed to query promocode: %w", err)
	}

	if planID.Valid {
		val := int(planID.Int32)
		pc.PlanID = &val
	}
	if subID.Valid {
		val := int(subID.Int32)
		pc.SubID = &val
	}
	if expiresAt.Valid {
		pc.ExpiresAt = expiresAt.Time
	}
	if createdAt.Valid {
		pc.CreatedAt = createdAt.Time
	}

	statusObj, err := domain.NewPromocodeStatus(status)
	if err != nil {
		return domain.Promocode{}, err
	}
	pc.Status = statusObj

	return pc, nil
}

func (r *PromocodeRepo) GetByService(ctx context.Context, serviceID int) ([]domain.Promocode, error) {
	rows, err := r.p.Query(ctx, getPromocodesByService, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query promocodes: %w", err)
	}
	defer rows.Close()

	var promocodes []domain.Promocode
	for rows.Next() {
		var pc domain.Promocode
		var planID pgtype.Int4
		var subID pgtype.Int4
		var expiresAt pgtype.Date
		var createdAt pgtype.Timestamp
		var status string

		err := rows.Scan(
			&pc.PromocodeID,
			&pc.ServiceID,
			&pc.Value,
			&planID,
			&subID,
			&expiresAt,
			&createdAt,
			&pc.Discount,
			&pc.MaxUses,
			&pc.CurUses,
			&status,
			&pc.DurationDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan promocode: %w", err)
		}

		if planID.Valid {
			val := int(planID.Int32)
			pc.PlanID = &val
		}
		if subID.Valid {
			val := int(subID.Int32)
			pc.SubID = &val
		}
		if expiresAt.Valid {
			pc.ExpiresAt = expiresAt.Time
		}
		if createdAt.Valid {
			pc.CreatedAt = createdAt.Time
		}

		statusObj, err := domain.NewPromocodeStatus(status)
		if err != nil {
			return nil, err
		}
		pc.Status = statusObj

		promocodes = append(promocodes, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return promocodes, nil
}

func (r *PromocodeRepo) GetAll(ctx context.Context) ([]domain.Promocode, error) {
	rows, err := r.p.Query(ctx, getAllPromocodes)
	if err != nil {
		return nil, fmt.Errorf("failed to query promocodes: %w", err)
	}
	defer rows.Close()

	var promocodes []domain.Promocode
	for rows.Next() {
		var pc domain.Promocode
		var planID pgtype.Int4
		var subID pgtype.Int4
		var expiresAt pgtype.Date
		var createdAt pgtype.Timestamp
		var status string

		err := rows.Scan(
			&pc.PromocodeID,
			&pc.ServiceID,
			&pc.Value,
			&planID,
			&subID,
			&expiresAt,
			&createdAt,
			&pc.Discount,
			&pc.MaxUses,
			&pc.CurUses,
			&status,
			&pc.DurationDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan promocode: %w", err)
		}

		if planID.Valid {
			val := int(planID.Int32)
			pc.PlanID = &val
		}
		if subID.Valid {
			val := int(subID.Int32)
			pc.SubID = &val
		}
		if expiresAt.Valid {
			pc.ExpiresAt = expiresAt.Time
		}
		if createdAt.Valid {
			pc.CreatedAt = createdAt.Time
		}

		statusObj, err := domain.NewPromocodeStatus(status)
		if err != nil {
			return nil, err
		}
		pc.Status = statusObj

		promocodes = append(promocodes, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return promocodes, nil
}

func (r *PromocodeRepo) Create(ctx context.Context, pc domain.Promocode) (domain.PromocodeID, error) {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var planID any = nil
	if pc.PlanID != nil {
		planID = *pc.PlanID
	}
	var subID any = nil
	if pc.SubID != nil {
		subID = *pc.SubID
	}
	// Ensure created_at is set
	if pc.CreatedAt.IsZero() {
		pc.CreatedAt = time.Now()
	}

	var id int
	err = tx.QueryRow(ctx, insertPromocode,
		pc.ServiceID,
		pc.Value,
		planID,
		subID,
		pc.ExpiresAt,
		pc.CreatedAt,
		pc.Discount,
		pc.MaxUses,
		pc.CurUses,
		pc.Status.String(),
		pc.DurationDays,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert promocode: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.PromocodeID(id), nil
}

func (r *PromocodeRepo) Update(ctx context.Context, pc domain.Promocode) error {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var planID any = nil
	if pc.PlanID != nil {
		planID = *pc.PlanID
	}
	var subID any = nil
	if pc.SubID != nil {
		subID = *pc.SubID
	}

	_, err = tx.Exec(ctx, updatePromocode,
		int(pc.PromocodeID),
		pc.ServiceID,
		pc.Value,
		planID,
		subID,
		pc.ExpiresAt,
		pc.CreatedAt,
		pc.Discount,
		pc.MaxUses,
		pc.CurUses,
		pc.Status.String(),
		pc.DurationDays,
	)
	if err != nil {
		return fmt.Errorf("failed to update promocode: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PromocodeRepo) Delete(ctx context.Context, id domain.PromocodeID) error {
	res, err := r.p.Exec(ctx, deletePromocode, int(id))
	if err != nil {
		return err
	}
	if r := res.RowsAffected(); r == 0 {
		return domain.ErrNoPromocodesDeleted
	}
	return nil
}

func (r *PromocodeRepo) IncrementUses(ctx context.Context, id domain.PromocodeID) error {
	_, err := r.p.Exec(ctx, incrementUses, int(id))
	if err != nil {
		return fmt.Errorf("failed to increment uses: %w", err)
	}
	return nil
}
