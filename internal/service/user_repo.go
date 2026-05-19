package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type UserRepo struct {
	p *pgxpool.Pool
}

func NewUserRepo(p *pgxpool.Pool) (domain.UserRepository, error) {
	if p == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	return &UserRepo{p: p}, nil
}

const (
	storeUserQuery = `
INSERT INTO users (user_id, email, password, user_name, age, balance, referral_code, role)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
	email = EXCLUDED.email,
	password = EXCLUDED.password,
	user_name = EXCLUDED.user_name,
	age = EXCLUDED.age,
	balance = EXCLUDED.balance,
	referral_code = EXCLUDED.referral_code,
	role = EXCLUDED.role;
`
	getUserQuery = `
SELECT user_id, email, password, user_name, age, balance, referral_code, role
FROM users
WHERE user_id = $1;
`
	updateUserQuery = `
UPDATE users
SET email = $2, password = $3, user_name = $4, age = $5, balance = $6, referral_code = $7, role = $8
WHERE user_id = $1;
`
	getUserByReferralCodeQuery = `
SELECT user_id, email, password, user_name, age, balance, referral_code, role
FROM users
WHERE referral_code = $1;
`
	storeReferralQuery = `
INSERT INTO user_referrals (referrer_id, referred_id)
VALUES ($1, $2);
`
	setAppCurrentUserIDQuery = `
SELECT set_config('app.current_user_id', $1, false);
`
)

func (r *UserRepo) StoreUser(ctx context.Context, user domain.User) error {
	_, err := r.p.Exec(ctx, storeUserQuery,
		user.UserID,
		user.Email,
		user.Password,
		user.UserName,
		user.Age,
		user.Balance,
		user.ReferralCode,
		user.Role,
	)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	var user domain.User
	var referralCode *string
	var role string
	err := r.p.QueryRow(ctx, getUserQuery, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.Password,
		&user.UserName,
		&user.Age,
		&user.Balance,
		&referralCode,
		&role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	user.ReferralCode = referralCode
	user.Role = domain.Role(role)
	// CreatedAt and UpdatedAt are not stored, set to zero values
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	return user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user domain.User) error {
	result, err := r.p.Exec(ctx, updateUserQuery,
		user.UserID,
		user.Email,
		user.Password,
		user.UserName,
		user.Age,
		user.Balance,
		user.ReferralCode,
		user.Role,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) GetUserByReferralCode(ctx context.Context, referralCode string) (domain.User, error) {
	var user domain.User
	var refCode *string
	var role string
	err := r.p.QueryRow(ctx, getUserByReferralCodeQuery, referralCode).Scan(
		&user.UserID,
		&user.Email,
		&user.Password,
		&user.UserName,
		&user.Age,
		&user.Balance,
		&refCode,
		&role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("failed to get user by referral code: %w", err)
	}
	user.ReferralCode = refCode
	user.Role = domain.Role(role)
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	return user, nil
}

func (r *UserRepo) StoreReferral(ctx context.Context, referrerID, referredID uuid.UUID) error {
	_, err := r.p.Exec(ctx, storeReferralQuery, referrerID, referredID)
	if err != nil {
		log.Printf("failed to store referral: %v", err)
		return fmt.Errorf("failed to store referral: %w", err)
	}
	return nil
}

func (r *UserRepo) SetAppCurrentUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.p.Exec(ctx, setAppCurrentUserIDQuery, userID.String())
	if err != nil {
		log.Printf("failed to set app.current_user_id: %v", err)
		return fmt.Errorf("failed to set app.current_user_id: %w", err)
	}
	return nil
}