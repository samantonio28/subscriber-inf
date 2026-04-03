package domain

import "errors"

var (
	ErrInvalidSubRepo     = errors.New("subscription repository not defined")
	ErrInvalidSubPlanRepo = errors.New("subscription plan repository not defined")
	ErrInvalidLogger      = errors.New("logger is not defined")
	ErrInvalidInput       = errors.New("invalid input")

	ErrPromocodeNotFound        = errors.New("promocode not found")
	ErrSubscriptionPlanNotFound = errors.New("subscription plan not found")
	ErrSubscriptionNotFound     = errors.New("subscription not found")

	// Deletion errors
	ErrNoPromocodesDeleted   = errors.New("no promocodes deleted")
	ErrNoSubscriptionDeleted = errors.New("no subscription deleted")
	ErrNoPlansDeleted        = errors.New("no plans deleted")
)
