package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/subscription_plan_repository.go -package=mocks . SubscriptionPlanRepository
//go:generate mockgen -destination=../mocks/subscription_plan_cache.go -package=mocks . SubscriptionPlanCache

type SubscriptionPlanRepository interface {
	GetByID(ctx context.Context, id PlanID) (SubscriptionPlan, error)
	GetByService(ctx context.Context, serviceID int) ([]SubscriptionPlan, error)
	GetAll(ctx context.Context) ([]SubscriptionPlan, error)
	Create(ctx context.Context, plan SubscriptionPlan) (PlanID, error)
	Update(ctx context.Context, plan SubscriptionPlan) error
	Delete(ctx context.Context, id PlanID) error
}

type SubscriptionPlanCache interface {
	GetSubscriptionPlan(ctx context.Context, planID string) (SubscriptionPlan, error)
	SetSubscriptionPlan(ctx context.Context, planID string, plan SubscriptionPlan, ttl time.Duration) error
	DeleteSubscriptionPlan(ctx context.Context, planID string) error
}