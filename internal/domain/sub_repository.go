package domain

import (
	"context"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Sub(ctx context.Context, subId SubID) (Subscription, error)
	UserSubs(ctx context.Context, userId uuid.UUID) ([]Subscription, error)
	StoreSub(ctx context.Context, sub Subscription) (SubID, error)
	UpdateSub(ctx context.Context, sub Subscription) error
	DeleteSub(ctx context.Context, subId SubID) error
	SubsTotalCosts(ctx context.Context, filter SubsFilter) (int, []SubID, error)
}
