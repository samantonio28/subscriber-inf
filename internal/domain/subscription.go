package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserID uuid.UUID
type SubID int

type Subscription struct {
	SubId       SubID
	UserID      UserID
	ServiceName string
	Price       int
	StartDate   time.Time
	EndDate     time.Time
}

func NewSubscription(subId SubID, userID UserID, serviceName string, price int, startDate time.Time, endDate time.Time) (*Subscription, error) {
	if subId < 0 {
		return nil, errors.New("subId must be greater than 0")
	}
	if serviceName == "" {
		return nil, errors.New("serviceName must not be empty")
	}
	if price < 0 {
		return nil, errors.New("price must be greater than 0")
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
		Price:       price,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type SubsFilter struct {
	StartDate   time.Time
	EndDate     time.Time
	UserID      UserID
	ServiceName string
}

func NewSubsFilter(startDate time.Time, endDate time.Time, userID UserID, serviceName string) (*SubsFilter, error) {
	if startDate.IsZero() {
		return nil, errors.New("startDate must not be zero")
	}
	if !endDate.IsZero() && endDate.Before(startDate) {
		return nil, errors.New("endDate must be greater than startDate")
	}
	return &SubsFilter{
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      userID,
		ServiceName: serviceName,
	}, nil
}
