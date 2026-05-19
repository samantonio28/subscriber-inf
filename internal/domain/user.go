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
	Role         Role
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
	role Role,
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
	if !role.Valid() {
		role = RoleUser
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
		Role:         role,
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
	SetAppCurrentUserID(ctx context.Context, userID uuid.UUID) error
}

