package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) (domain.UserRepository, error) {
	if db == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	return &UserRepo{db: db}, nil
}

const (
	storeUserQuery = `
INSERT INTO users (user_id, email, password, user_name, age, balance, referral_code)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	email = VALUES(email),
	password = VALUES(password),
	user_name = VALUES(user_name),
	age = VALUES(age),
	balance = VALUES(balance),
	referral_code = VALUES(referral_code);
`
	getUserQuery = `
SELECT user_id, email, password, user_name, age, balance, referral_code
FROM users
WHERE user_id = ?;
`
	updateUserQuery = `
UPDATE users
SET email = ?, password = ?, user_name = ?, age = ?, balance = ?, referral_code = ?
WHERE user_id = ?;
`
	getUserByReferralCodeQuery = `
SELECT user_id, email, password, user_name, age, balance, referral_code
FROM users
WHERE referral_code = ?;
`
	storeReferralQuery = `
INSERT INTO user_referrals (referrer_id, referred_id)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE referred_id = VALUES(referred_id);
`
)

func (r *UserRepo) StoreUser(ctx context.Context, user domain.User) error {
	_, err := r.db.ExecContext(ctx, storeUserQuery,
		user.UserID,
		user.Email,
		user.Password,
		user.UserName,
		user.Age,
		user.Balance,
		user.ReferralCode,
	)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	var user domain.User
	var referralCode sql.NullString
	err := r.db.QueryRowContext(ctx, getUserQuery, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.Password,
		&user.UserName,
		&user.Age,
		&user.Balance,
		&referralCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	if referralCode.Valid {
		user.ReferralCode = &referralCode.String
	} else {
		user.ReferralCode = nil
	}
	// CreatedAt and UpdatedAt are not stored, set to zero values
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	return user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	_, err = tx.ExecContext(ctx, updateUserQuery,
		user.Email,
		user.Password,
		user.UserName,
		user.Age,
		user.Balance,
		user.ReferralCode,
		user.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *UserRepo) GetUserByReferralCode(ctx context.Context, referralCode string) (domain.User, error) {
	var user domain.User
	var refCode sql.NullString
	err := r.db.QueryRowContext(ctx, getUserByReferralCodeQuery, referralCode).Scan(
		&user.UserID,
		&user.Email,
		&user.Password,
		&user.UserName,
		&user.Age,
		&user.Balance,
		&refCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("failed to get user by referral code: %w", err)
	}
	if refCode.Valid {
		user.ReferralCode = &refCode.String
	} else {
		user.ReferralCode = nil
	}
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	return user, nil
}

func (r *UserRepo) StoreReferral(ctx context.Context, referrerID, referredID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, storeReferralQuery, referrerID, referredID)
	if err != nil {
		return fmt.Errorf("failed to store referral: %w", err)
	}
	return nil
}