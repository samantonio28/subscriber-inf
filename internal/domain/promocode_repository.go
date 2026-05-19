package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/promocode_repository.go -package=mocks . PromocodeRepository
//go:generate mockgen -destination=../mocks/promocode_cache.go -package=mocks . PromocodeCache

type PromocodeRepository interface {
	GetByID(ctx context.Context, id PromocodeID) (Promocode, error)
	GetByCode(ctx context.Context, code string) (Promocode, error)
	GetByService(ctx context.Context, serviceID int) ([]Promocode, error)
	GetAll(ctx context.Context) ([]Promocode, error)
	Create(ctx context.Context, pc Promocode) (PromocodeID, error)
	Update(ctx context.Context, pc Promocode) error
	Delete(ctx context.Context, id PromocodeID) error
	IncrementUses(ctx context.Context, id PromocodeID) error
}

type PromocodeCache interface {
	GetPromocode(ctx context.Context, promoID string) (Promocode, error)
	SetPromocode(ctx context.Context, promoID string, promo Promocode, ttl time.Duration) error
	DeletePromocode(ctx context.Context, promoID string) error
}