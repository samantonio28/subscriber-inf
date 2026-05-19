package delivery

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	userRoleContextKey contextKey = "user_role"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

func WithUserRole(ctx context.Context, role domain.Role) context.Context {
	return context.WithValue(ctx, userRoleContextKey, role)
}

func UserRoleFromContext(ctx context.Context) (domain.Role, bool) {
	val := ctx.Value(userRoleContextKey)
	if val == nil {
		return "", false
	}
	role, ok := val.(domain.Role)
	return role, ok
}
