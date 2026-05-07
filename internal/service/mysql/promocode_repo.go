package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type PromocodeRepo struct {
	db *sql.DB
}

func NewPromocodeRepo(db *sql.DB) (domain.PromocodeRepository, error) {
	if db == nil {
		return nil, domain.ErrInvalidPromoRepo
	}
	return &PromocodeRepo{db: db}, nil
}

const (
	getPromocodeByID = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE promocode_id = ?;
`
	getPromocodeByCode = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE promocode = ?;
`
	getPromocodesByService = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes
WHERE service_id = ?;
`
	insertPromocode = `
INSERT INTO promocodes (service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	updatePromocode = `
UPDATE promocodes
SET service_id = ?, promocode = ?, plan_id = ?, sub_id = ?, expires_at = ?, created_at = ?, discount = ?, max_uses = ?, cur_uses = ?, status = ?, duration_days = ?
WHERE promocode_id = ?;
`
	deletePromocode = `
DELETE FROM promocodes WHERE promocode_id = ?;
`
	incrementUses = `
UPDATE promocodes
SET cur_uses = cur_uses + 1
WHERE promocode_id = ?;
`
	getAllPromocodes = `
SELECT promocode_id, service_id, promocode, plan_id, sub_id, expires_at, created_at, discount, max_uses, cur_uses, status, duration_days
FROM promocodes;
`
)

func (r *PromocodeRepo) GetByID(ctx context.Context, id domain.PromocodeID) (domain.Promocode, error) {
	var pc domain.Promocode
	var planID sql.NullInt32
	var subID sql.NullInt32
	var expiresAt sql.NullTime
	var createdAt sql.NullTime
	var status string

	err := r.db.QueryRowContext(ctx, getPromocodeByID, int(id)).Scan(
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
	var planID sql.NullInt32
	var subID sql.NullInt32
	var expiresAt sql.NullTime
	var createdAt sql.NullTime
	var status string

	err := r.db.QueryRowContext(ctx, getPromocodeByCode, code).Scan(
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
	rows, err := r.db.QueryContext(ctx, getPromocodesByService, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query promocodes by service: %w", err)
	}
	defer rows.Close()

	var promocodes []domain.Promocode
	for rows.Next() {
		var pc domain.Promocode
		var planID sql.NullInt32
		var subID sql.NullInt32
		var expiresAt sql.NullTime
		var createdAt sql.NullTime
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
			return nil, fmt.Errorf("failed to scan promocode row: %w", err)
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
		return nil, fmt.Errorf("error iterating promocode rows: %w", err)
	}
	return promocodes, nil
}

func (r *PromocodeRepo) Create(ctx context.Context, pc domain.Promocode) (domain.PromocodeID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var planIDOrNil any = pc.PlanID
	if pc.PlanID == nil {
		planIDOrNil = nil
	}
	var subIDOrNil any = pc.SubID
	if pc.SubID == nil {
		subIDOrNil = nil
	}
	var expiresAtOrNil any = pc.ExpiresAt
	if pc.ExpiresAt.IsZero() {
		expiresAtOrNil = nil
	}
	var createdAtOrNil any = pc.CreatedAt
	if pc.CreatedAt.IsZero() {
		createdAtOrNil = time.Now()
	}

	result, err := tx.ExecContext(ctx, insertPromocode,
		pc.ServiceID,
		pc.Value,
		planIDOrNil,
		subIDOrNil,
		expiresAtOrNil,
		createdAtOrNil,
		pc.Discount,
		pc.MaxUses,
		pc.CurUses,
		pc.Status.String(),
		pc.DurationDays,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert promocode: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return domain.PromocodeID(id), nil
}

func (r *PromocodeRepo) Update(ctx context.Context, pc domain.Promocode) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var planIDOrNil any = pc.PlanID
	if pc.PlanID == nil {
		planIDOrNil = nil
	}
	var subIDOrNil any = pc.SubID
	if pc.SubID == nil {
		subIDOrNil = nil
	}
	var expiresAtOrNil any = pc.ExpiresAt
	if pc.ExpiresAt.IsZero() {
		expiresAtOrNil = nil
	}
	var createdAtOrNil any = pc.CreatedAt
	if pc.CreatedAt.IsZero() {
		createdAtOrNil = time.Now()
	}

	_, err = tx.ExecContext(ctx, updatePromocode,
		pc.ServiceID,
		pc.Value,
		planIDOrNil,
		subIDOrNil,
		expiresAtOrNil,
		createdAtOrNil,
		pc.Discount,
		pc.MaxUses,
		pc.CurUses,
		pc.Status.String(),
		pc.DurationDays,
		int(pc.PromocodeID),
	)
	if err != nil {
		return fmt.Errorf("failed to update promocode: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PromocodeRepo) Delete(ctx context.Context, id domain.PromocodeID) error {
	res, err := r.db.ExecContext(ctx, deletePromocode, int(id))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNoPromocodesDeleted
	}
	return nil
}

func (r *PromocodeRepo) IncrementUses(ctx context.Context, id domain.PromocodeID) error {
	_, err := r.db.ExecContext(ctx, incrementUses, int(id))
	return err
}

func (r *PromocodeRepo) GetAll(ctx context.Context) ([]domain.Promocode, error) {
	rows, err := r.db.QueryContext(ctx, getAllPromocodes)
	if err != nil {
		return nil, fmt.Errorf("failed to query all promocodes: %w", err)
	}
	defer rows.Close()

	var promocodes []domain.Promocode
	for rows.Next() {
		var pc domain.Promocode
		var planID sql.NullInt32
		var subID sql.NullInt32
		var expiresAt sql.NullTime
		var createdAt sql.NullTime
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
			return nil, fmt.Errorf("failed to scan promocode row: %w", err)
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
		return nil, fmt.Errorf("error iterating promocode rows: %w", err)
	}
	return promocodes, nil
}