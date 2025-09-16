package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SubID int
type SubscriptionType struct {
	s string
}

var (
	Usual     = SubscriptionType{"usual"}
	Promocode = SubscriptionType{"promocode"}
	Family    = SubscriptionType{"family"}
)

func (t *SubscriptionType) String() string {
	return t.s
}

func NewSubscriptionType(s string) (*SubscriptionType, error) {
	switch s {
	case "usual", "":
		return &Usual, nil
	case "promocode":
		return &Promocode, nil
	case "family":
		return &Family, nil
	default:
		return nil, fmt.Errorf(
			"%s: bad subscription type: not matches 'usual', 'promocode', 'family', got %s",
			ErrInvalidInput, s,
		)
	}
}

type Subscription struct {
	SubId       SubID
	UserID      uuid.UUID
	ServiceName string
	Price       int
	SubType     SubscriptionType
	StartDate   time.Time
	EndDate     time.Time
}

func NewSubscription(
	subId SubID,
	userID uuid.UUID,
	serviceName string,
	price int,
	subType string,
	startDate time.Time,
	endDate time.Time,
) (*Subscription, error) {
	if subId < 0 {
		return nil, errors.New("subId must be greater than 0")
	}
	if serviceName == "" {
		return nil, errors.New("serviceName must not be empty")
	}
	if price < 0 {
		return nil, errors.New("price must be greater than 0")
	}
	subtype, err := NewSubscriptionType(subType)
	if err != nil {
		return nil, err
	}
	if startDate.IsZero() {
		return nil, errors.New("startDate must not be zero")
	}
	if !endDate.IsZero() && endDate.Before(startDate) {
		return nil, errors.New("endDate must be greater than startDate")
	}
	return &Subscription{
		SubId:       subId,
		UserID:      userID,
		ServiceName: serviceName,
		SubType:     *subtype,
		Price:       price,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type SubsFilter struct {
	StartDate   time.Time
	EndDate     time.Time
	UserID      uuid.UUID
	ServiceName string
	SubType     SubscriptionType
}

func NewSubsFilter(startDate time.Time, endDate time.Time, userID uuid.UUID, serviceName string, subType string) (*SubsFilter, error) {
	if startDate.IsZero() {
		return nil, errors.New("startDate must not be zero")
	}
	if !endDate.IsZero() && endDate.Before(startDate) {
		return nil, errors.New("endDate must be greater than startDate")
	}
	subtype, err := NewSubscriptionType(subType)
	if err != nil {
		return nil, err
	}
	return &SubsFilter{
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      userID,
		ServiceName: serviceName,
		SubType:     *subtype,
	}, nil
}
