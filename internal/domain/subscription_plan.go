package domain

import (
	"errors"
)

type PlanID int

type SubscriptionPlan struct {
	PlanID       PlanID
	ServiceID    int
	Name         string
	DurationDays int
	Price        int
}

func NewSubscriptionPlan(
	planID PlanID,
	serviceID int,
	name string,
	durationDays int,
	price int,
) (*SubscriptionPlan, error) {
	if serviceID <= 0 {
		return nil, errors.New("serviceID must be greater than 0")
	}
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	if durationDays <= 0 {
		return nil, errors.New("durationDays must be greater than 0")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	return &SubscriptionPlan{
		PlanID:       planID,
		ServiceID:    serviceID,
		Name:         name,
		DurationDays: durationDays,
		Price:        price,
	}, nil
}