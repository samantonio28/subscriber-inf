package domain

import (
	"context"
)

type SubscriptionRepository interface {
	Sub(ctx context.Context, subId SubID) (*Subscription, error)
	UserSubs(ctx context.Context, userId UserID) ([]*Subscription, error)
	StoreSub(ctx context.Context, sub Subscription) error
	UpdateSub(ctx context.Context, sub Subscription) error
	DeleteSub(ctx context.Context, subId SubID) error
	SubsTotalCosts(ctx context.Context, filter SubsFilter) (int, []SubID, error)
}
