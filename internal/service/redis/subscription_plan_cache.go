package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/redis"
)

type SubscriptionPlanCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSubscriptionPlanCache(client *redis.Client) (*SubscriptionPlanCache, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &SubscriptionPlanCache{
		client: client,
		ttl:    10 * time.Minute,
	}, nil
}

func (c *SubscriptionPlanCache) GetSubscriptionPlan(ctx context.Context, planID string) (domain.SubscriptionPlan, error) {
	key := subscriptionPlanCacheKey(planID)
	val, err := c.client.Get(ctx, key)
	if err != nil {
		return domain.SubscriptionPlan{}, err
	}
	var plan domain.SubscriptionPlan
	if err := json.Unmarshal([]byte(val), &plan); err != nil {
		return domain.SubscriptionPlan{}, err
	}
	return plan, nil
}

func (c *SubscriptionPlanCache) SetSubscriptionPlan(ctx context.Context, planID string, plan domain.SubscriptionPlan, ttl time.Duration) error {
	key := subscriptionPlanCacheKey(planID)
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	if ttl == 0 {
		ttl = c.ttl
	}
	return c.client.Set(ctx, key, data, ttl)
}

func (c *SubscriptionPlanCache) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	key := subscriptionPlanCacheKey(planID)
	return c.client.Delete(ctx, key)
}

func subscriptionPlanCacheKey(planID string) string {
	return fmt.Sprintf("subscription_plan:%s", planID)
}

// Helper to convert PlanID to string
func planIDToString(planID domain.PlanID) string {
	return strconv.Itoa(int(planID))
}