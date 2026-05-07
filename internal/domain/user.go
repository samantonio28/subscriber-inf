package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID
	Email        string
	Password     string
	UserName     string
	Age          int
	Balance      int
	ReferralCode *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	userID uuid.UUID,
	email string,
	password string,
	userName string,
	age int,
	balance int,
	referralCode *string,
) (*User, error) {
	if email == "" {
		return nil, errors.New("email must not be empty")
	}
	if password == "" {
		return nil, errors.New("password must not be empty")
	}
	if userName == "" {
		return nil, errors.New("userName must not be empty")
	}
	if age < 0 {
		return nil, errors.New("age must be greater than or equal to 0")
	}
	if balance < 0 {
		return nil, errors.New("balance must be greater than or equal to 0")
	}
	if referralCode == nil {
		referralCode = generateReferralCode()
	}
	return &User{
		UserID:       userID,
		Email:        email,
		Password:     password,
		UserName:     userName,
		Age:          age,
		Balance:      balance,
		ReferralCode: referralCode,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func generateReferralCode() *string {
	code := uuid.New().String()[:8]
	return &code
}

type UserRepository interface {
	StoreUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, user User) error
	GetUserByReferralCode(ctx context.Context, referralCode string) (User, error)
	StoreReferral(ctx context.Context, referrerID, referredID uuid.UUID) error
}
