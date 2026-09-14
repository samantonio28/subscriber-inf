package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

// TestUUID возвращает новый UUID (единая точка генерации для тестов).
func TestUUID() uuid.UUID { return uuid.New() }

// TestUser возвращает пользователя с заданным балансом и валидными полями.
func TestUser(balance int) domain.User {
	code := "REFCODE1"
	return domain.User{
		UserID:       TestUUID(),
		Email:        "user@example.com",
		Password:     "secret",
		UserName:     "Test User",
		Age:          25,
		Balance:      balance,
		ReferralCode: &code,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestSubscription возвращает активную обычную подписку.
func TestSubscription() domain.Subscription {
	now := time.Now()
	return domain.Subscription{
		SubId:       1,
		UserID:      TestUUID(),
		ServiceName: "Netflix",
		Price:       1000,
		SubType:     domain.SubTypeUsual,
		StartDate:   now.Add(-24 * time.Hour),
		EndDate:     now.Add(30 * 24 * time.Hour),
		PlanID:      1,
	}
}

// TestPromocode возвращает активный промокод с заданной скидкой.
func TestPromocode(discount, maxUses int) domain.Promocode {
	return domain.Promocode{
		PromocodeID:  1,
		ServiceID:    1,
		Value:        "SUMMER2026",
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		Discount:     discount,
		MaxUses:      maxUses,
		CurUses:      0,
		Status:       domain.PromocodeStatusActive,
		DurationDays: 30,
	}
}
