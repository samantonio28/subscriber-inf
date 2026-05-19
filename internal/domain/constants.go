package domain

import "errors"

var (
	ErrInvalidSubRepo         = errors.New("subscription repository not defined")
	ErrInvalidSubPlanRepo     = errors.New("subscription plan repository not defined")
	ErrInvalidPromoRepo       = errors.New("promocode repository not defined")
	ErrInvalidUserRepo        = errors.New("user repository not defined")
	ErrInvalidPaymentRepo     = errors.New("payment repository not defined")
	ErrInvalidStatsService    = errors.New("stats service not defined")
	ErrInvalidUserServiceRepo = errors.New("user service repository not defined")
	ErrInvalidLogger          = errors.New("logger is not defined")
	ErrInvalidCache           = errors.New("cache is not defined")
	ErrInvalidInput           = errors.New("invalid input")

	ErrUserNotFound             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrPromocodeNotFound        = errors.New("promocode not found")
	ErrSubscriptionPlanNotFound = errors.New("subscription plan not found")
	ErrSubscriptionNotFound     = errors.New("subscription not found")

	// Deletion errors
	ErrNoPromocodesDeleted   = errors.New("no promocodes deleted")
	ErrNoSubscriptionDeleted = errors.New("no subscription deleted")
	ErrNoPlansDeleted        = errors.New("no plans deleted")

	// Promocode validation errors
	ErrPromocodeNotActive              = errors.New("promocode is not active")
	ErrPromocodeExpired                = errors.New("promocode expired")
	ErrPromocodeMaxUsesReached         = errors.New("promocode max uses reached")
	ErrPromocodeNotApplicable          = errors.New("promocode not applicable to this subscription")
	ErrPromocodeAlreadyApplied         = errors.New("promocode already applied to this subscription")
	ErrSubscriptionAlreadyHasPromocode = errors.New("subscription already has promocode type")
	ErrPromocodeNotForNewSubscription  = errors.New("promocode cannot be applied to new subscription")
	ErrPromocodeNotForRenewal          = errors.New("promocode cannot be applied to renewal")

	// Payment errors
	ErrInsufficientBalance = errors.New("insufficient balance")
)
