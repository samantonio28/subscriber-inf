package domain

import (
	"errors"
	"fmt"
	"time"
)

type PromocodeID int

type PromocodeStatus string

const (
	PromocodeStatusActive   PromocodeStatus = "ACTIVE"
	PromocodeStatusUsed     PromocodeStatus = "USED"
	PromocodeStatusDisabled PromocodeStatus = "DISABLED"
)

func (ps PromocodeStatus) String() string {
	return string(ps)
}

func NewPromocodeStatus(s string) (PromocodeStatus, error) {
	switch s {
	case "ACTIVE":
		return PromocodeStatusActive, nil
	case "USED":
		return PromocodeStatusUsed, nil
	case "DISABLED":
		return PromocodeStatusDisabled, nil
	default:
		return "", fmt.Errorf(
			"%s: bad promocode status: not matches 'ACTIVE', 'USED', 'DISABLED', got %s",
			ErrInvalidInput, s,
		)
	}
}

type Promocode struct {
	PromocodeID  PromocodeID
	ServiceID    int
	Value        string
	PlanID       *int // nullable, reference to subscription_plans
	SubID        *int // nullable, reference to subscriptions (for backward compatibility)
	ExpiresAt    time.Time
	CreatedAt    time.Time
	Discount     int // percentage 0-100
	MaxUses      int
	CurUses      int
	Status       PromocodeStatus
	DurationDays int // optional, default 3 days
}

func NewPromocode(
	promocodeID PromocodeID,
	serviceID int,
	value string,
	planID *int,
	subID *int,
	expiresAt time.Time,
	createdAt time.Time,
	discount int,
	maxUses int,
	curUses int,
	status PromocodeStatus,
	durationDays int,
) (*Promocode, error) {
	if serviceID <= 0 {
		return nil, errors.New("serviceID must be greater than 0")
	}
	if value == "" {
		return nil, errors.New("promocode must not be empty")
	}
	if discount < 0 || discount > 100 {
		return nil, errors.New("discount must be between 0 and 100")
	}
	if maxUses < 1 {
		return nil, errors.New("maxUses must be at least 1")
	}
	if curUses < 0 || curUses > maxUses {
		return nil, errors.New("curUses must be between 0 and maxUses")
	}
	if expiresAt.IsZero() {
		return nil, errors.New("expiresAt must not be zero")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	if durationDays <= 0 {
		durationDays = 3
	}
	return &Promocode{
		PromocodeID:  promocodeID,
		ServiceID:    serviceID,
		Value:        value,
		PlanID:       planID,
		SubID:        subID,
		ExpiresAt:    expiresAt,
		CreatedAt:    createdAt,
		Discount:     discount,
		MaxUses:      maxUses,
		CurUses:      curUses,
		Status:       status,
		DurationDays: durationDays,
	}, nil
}
