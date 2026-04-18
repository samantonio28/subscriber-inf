package domain

import (
	"context"
)

type SubscriptionPlanRepository interface {
	GetByID(ctx context.Context, id PlanID) (SubscriptionPlan, error)
	GetByService(ctx context.Context, serviceID int) ([]SubscriptionPlan, error)
	GetAll(ctx context.Context) ([]SubscriptionPlan, error)
	Create(ctx context.Context, plan SubscriptionPlan) (PlanID, error)
	Update(ctx context.Context, plan SubscriptionPlan) error
	Delete(ctx context.Context, id PlanID) error
}