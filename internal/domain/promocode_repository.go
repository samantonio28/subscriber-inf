package domain

import (
	"context"
)

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