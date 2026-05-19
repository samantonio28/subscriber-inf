package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserServiceRepository interface {
	Add(ctx context.Context, userID uuid.UUID, serviceID int) error
	Remove(ctx context.Context, userID uuid.UUID, serviceID int) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]int, error)
	ListByService(ctx context.Context, serviceID int) ([]uuid.UUID, error)
	Exists(ctx context.Context, userID uuid.UUID, serviceID int) (bool, error)
}
