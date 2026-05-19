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

const (
	minMaxUsesForCaching = 100
)

type PromocodeCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPromocodeCache(client *redis.Client) (*PromocodeCache, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &PromocodeCache{
		client: client,
		ttl:    10 * time.Minute,
	}, nil
}

func (c *PromocodeCache) GetPromocode(ctx context.Context, promoID string) (domain.Promocode, error) {
	key := promocodeCacheKey(promoID)
	val, err := c.client.Get(ctx, key)
	if err != nil {
		return domain.Promocode{}, err
	}
	var promo domain.Promocode
	if err := json.Unmarshal([]byte(val), &promo); err != nil {
		return domain.Promocode{}, err
	}
	return promo, nil
}

func (c *PromocodeCache) SetPromocode(ctx context.Context, promoID string, promo domain.Promocode, ttl time.Duration) error {
	// Кэшируем только если MaxUses >= minMaxUsesForCaching
	if promo.MaxUses < minMaxUsesForCaching {
		return nil
	}
	key := promocodeCacheKey(promoID)
	data, err := json.Marshal(promo)
	if err != nil {
		return err
	}
	if ttl == 0 {
		ttl = c.ttl
	}
	return c.client.Set(ctx, key, data, ttl)
}

func (c *PromocodeCache) DeletePromocode(ctx context.Context, promoID string) error {
	key := promocodeCacheKey(promoID)
	return c.client.Delete(ctx, key)
}

func promocodeCacheKey(promoID string) string {
	return fmt.Sprintf("promocode:%s", promoID)
}

// Helper to convert PromocodeID to string
func promocodeIDToString(promoID domain.PromocodeID) string {
	return strconv.Itoa(int(promoID))
}