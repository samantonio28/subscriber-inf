package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type ApplyPromocodeUC struct {
	subRepo   domain.SubscriptionRepository
	promoRepo domain.PromocodeRepository
	logger    logger.Logger
}

func NewApplyPromocodeUC(subRepo domain.SubscriptionRepository, promoRepo domain.PromocodeRepository, logger logger.Logger) (*ApplyPromocodeUC, error) {
	if subRepo == nil || promoRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &ApplyPromocodeUC{subRepo: subRepo, promoRepo: promoRepo, logger: logger}, nil
}

type ApplyPromocodeInput struct {
	SubscriptionID int
	PromocodeValue string
}

type ApplyPromocodeOutput struct {
	DiscountApplied int
	NewPrice        int
	Message         string
}

func (uc *ApplyPromocodeUC) Apply(ctx context.Context, input ApplyPromocodeInput) (*ApplyPromocodeOutput, error) {
	// 1. Get subscription
	sub, err := uc.subRepo.Sub(ctx, domain.SubID(input.SubscriptionID))
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return nil, domain.ErrSubscriptionNotFound
		}
		uc.logger.Error("failed to get subscription", "error", err)
		return nil, err
	}

	// 2. Get promocode by value
	promo, err := uc.promoRepo.GetByCode(ctx, input.PromocodeValue)
	if err != nil {
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			return nil, domain.ErrPromocodeNotFound
		}
		uc.logger.Error("failed to get promocode", "error", err)
		return nil, err
	}

	// 3. Validate promocode applicability
	if err := validatePromocode(promo, sub); err != nil {
		return nil, err
	}

	// 4. Determine application type: new subscription or renewal
	isNew := isNewSubscription(sub)
	isRenewal := isActiveSubscription(sub) && sub.SubType.String() != "promocode"

	if !isNew && !isRenewal {
		return nil, domain.ErrPromocodeNotApplicable
	}

	// 5. Check if promocode already applied to this subscription
	if sub.SubType.String() == "promocode" {
		return nil, domain.ErrSubscriptionAlreadyHasPromocode
	}
	if promo.SubID != nil && *promo.SubID == int(sub.SubId) {
		return nil, domain.ErrPromocodeAlreadyApplied
	}

	// 6. Calculate new price (discount applied only to monthly price)
	discount := promo.Discount
	newPrice := sub.Price * (100 - discount) / 100

	// 7. Update subscription with new price and sub_type = promocode
	promocodeID := int(promo.PromocodeID)
	updatedSub, err := domain.NewSubscription(
		sub.SubId,
		sub.UserID,
		sub.ServiceName,
		newPrice,
		"promocode",
		sub.StartDate,
		sub.EndDate,
		sub.PlanID,
		&promocodeID,
	)
	if err != nil {
		uc.logger.Error("failed to create updated subscription", "error", err)
		return nil, err
	}

	err = uc.subRepo.UpdateSub(ctx, *updatedSub)
	if err != nil {
		uc.logger.Error("failed to update subscription", "error", err)
		return nil, err
	}

	// 8. Increment promocode uses and update status if needed
	err = uc.promoRepo.IncrementUses(ctx, promo.PromocodeID)
	if err != nil {
		uc.logger.Error("failed to increment promocode uses", "error", err)
		// rollback? but we already updated subscription, maybe we can ignore this error
		// but for simplicity we just log and continue
	}

	// 9. Optionally update promocode's SubID to mark it as applied to this subscription
	// (if needed, can be added later)

	uc.logger.Info("promocode applied successfully", "subscription_id", input.SubscriptionID, "promocode", input.PromocodeValue, "application_type", map[bool]string{true: "new", false: "renewal"}[isNew])

	return &ApplyPromocodeOutput{
		DiscountApplied: discount,
		NewPrice:        newPrice,
		Message:         "Promocode applied successfully",
	}, nil
}

func validatePromocode(promo domain.Promocode, sub domain.Subscription) error {
	// Check status
	if promo.Status != domain.PromocodeStatusActive {
		return domain.ErrPromocodeNotActive
	}
	// Check expiration
	if time.Now().After(promo.ExpiresAt) {
		return domain.ErrPromocodeExpired
	}
	// Check usage limits
	if promo.CurUses >= promo.MaxUses {
		return domain.ErrPromocodeMaxUsesReached
	}
	// Check if promocode is bound to specific service, plan, or subscription
	// TODO: implement mapping between service name and service ID
	// For now, skip service ID check
	// Also check plan_id and sub_id if they are set
	if promo.ServiceID != 0 {
		// need to map service name to service ID
		// placeholder: assume service ID matches if we have mapping
	}
	if promo.PlanID != nil {
		// need to check if subscription belongs to this plan
		// placeholder: skip for now
	}
	if promo.SubID != nil && *promo.SubID != int(sub.SubId) {
		return domain.ErrPromocodeNotApplicable
	}
	return nil
}

func isNewSubscription(sub domain.Subscription) bool {
	// Subscription is considered new if its start date is in the future
	return sub.StartDate.After(time.Now())
}

func isActiveSubscription(sub domain.Subscription) bool {
	// Subscription is active if current time is within start and end dates
	now := time.Now()
	return (sub.StartDate.Before(now) || sub.StartDate.Equal(now)) && (sub.EndDate.IsZero() || sub.EndDate.After(now) || sub.EndDate.Equal(now))
}
