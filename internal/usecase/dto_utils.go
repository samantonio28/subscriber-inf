package usecase

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type SubscriptionDTO struct {
	SubId       int       `json:"sub_id"`
	UserId      uuid.UUID `json:"user_id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	SubType     string    `json:"sub_type"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	PlanID      int       `json:"plan_id"`
	PromocodeID *int      `json:"promocode_id,omitempty"`
}

type SubsFilterDTO struct {
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	ServiceName string    `json:"service_name,omitempty"`
	SubType     string    `json:"sub_type,omitempty"`
}

func SubToDTO(sub domain.Subscription) SubscriptionDTO {
	return SubscriptionDTO{
		SubId:       int(sub.SubId),
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		SubType:     sub.SubType.String(),
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
		PlanID:      sub.PlanID,
		PromocodeID: sub.PromocodeID,
	}
}

func DTOToSub(dto SubscriptionDTO) (domain.Subscription, error) {
	sub, err := domain.NewSubscription(
		domain.SubID(dto.SubId),
		dto.UserId,
		dto.ServiceName,
		dto.Price,
		dto.SubType,
		dto.StartDate,
		dto.EndDate,
		dto.PlanID,
		dto.PromocodeID,
	)
	if err != nil {
		return domain.Subscription{}, err
	}
	return *sub, nil
}

func FilterToDTO(fil domain.SubsFilter) SubsFilterDTO {
	return SubsFilterDTO{
		StartDate:   fil.StartDate,
		EndDate:     fil.EndDate,
		UserID:      fil.UserID,
		ServiceName: fil.ServiceName,
		SubType:     fil.SubType.String(),
	}
}

func DTOToFilter(dto SubsFilterDTO) (domain.SubsFilter, error) {
	f, err := domain.NewSubsFilter(
		dto.StartDate,
		dto.EndDate,
		dto.UserID,
		dto.ServiceName,
		dto.SubType,
	)
	if err != nil {
		return domain.SubsFilter{}, err
	}
	return *f, nil
}

type PromocodeDTO struct {
	PromocodeID  int       `json:"promocode_id"`
	ServiceID    int       `json:"service_id"`
	Value        string    `json:"value"`
	PlanID       *int      `json:"plan_id,omitempty"`
	SubID        *int      `json:"sub_id,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	Discount     int       `json:"discount"`
	MaxUses      int       `json:"max_uses"`
	CurUses      int       `json:"cur_uses"`
	Status       string    `json:"status"`
	DurationDays int       `json:"duration_days"`
}

func PromocodeToDTO(promo domain.Promocode) PromocodeDTO {
	return PromocodeDTO{
		PromocodeID:  int(promo.PromocodeID),
		ServiceID:    promo.ServiceID,
		Value:        promo.Value,
		PlanID:       promo.PlanID,
		SubID:        promo.SubID,
		ExpiresAt:    promo.ExpiresAt,
		CreatedAt:    promo.CreatedAt,
		Discount:     promo.Discount,
		MaxUses:      promo.MaxUses,
		CurUses:      promo.CurUses,
		Status:       string(promo.Status),
		DurationDays: promo.DurationDays,
	}
}

func DTOToPromocode(dto PromocodeDTO) (domain.Promocode, error) {
	status, err := domain.NewPromocodeStatus(dto.Status)
	if err != nil {
		return domain.Promocode{}, err
	}
	promo, err := domain.NewPromocode(
		domain.PromocodeID(dto.PromocodeID),
		dto.ServiceID,
		dto.Value,
		dto.PlanID,
		dto.SubID,
		dto.ExpiresAt,
		dto.CreatedAt,
		dto.Discount,
		dto.MaxUses,
		dto.CurUses,
		status,
		dto.DurationDays,
	)
	if err != nil {
		return domain.Promocode{}, err
	}
	return *promo, nil
}

type UserDTO struct {
	UserID       uuid.UUID   `json:"user_id"`
	Email        string      `json:"email"`
	Password     string      `json:"password"`
	UserName     string      `json:"user_name"`
	Age          int         `json:"age"`
	Balance      int         `json:"balance"`
	ReferralCode *string     `json:"referral_code"`
	Role         domain.Role `json:"role,omitempty"`
}

type GetUserDTO struct {
	UserID       uuid.UUID   `json:"user_id"`
	Email        string      `json:"email"`
	UserName     string      `json:"user_name"`
	Age          int         `json:"age"`
	Balance      int         `json:"balance"`
	ReferralCode *string     `json:"referral_code"`
	Role         domain.Role `json:"role,omitempty"`
}

func UserToDTO(user domain.User) UserDTO {
	return UserDTO{
		UserID:       user.UserID,
		Email:        user.Email,
		Password:     user.Password,
		UserName:     user.UserName,
		Age:          user.Age,
		Balance:      user.Balance,
		ReferralCode: user.ReferralCode,
		Role:         user.Role,
	}
}

func UserToGetUserDTO(user domain.User) GetUserDTO {
	return GetUserDTO{
		UserID:       user.UserID,
		Email:        user.Email,
		UserName:     user.UserName,
		Age:          user.Age,
		Balance:      user.Balance,
		ReferralCode: user.ReferralCode,
		Role:         user.Role,
	}
}

func DTOToUser(dto UserDTO) (domain.User, error) {
	role := dto.Role
	if !role.Valid() {
		role = domain.RoleUser
	}
	user, err := domain.NewUser(
		dto.UserID,
		dto.Email,
		dto.Password,
		dto.UserName,
		dto.Age,
		dto.Balance,
		dto.ReferralCode,
		role,
	)
	if err != nil {
		return domain.User{}, err
	}
	return *user, nil
}

type SubscriptionPlanDTO struct {
	PlanID       int    `json:"plan_id"`
	ServiceID    int    `json:"service_id"`
	Name         string `json:"name"`
	ServiceName  string `json:"service_name"`
	SubType      string `json:"sub_type"`
	DurationDays int    `json:"duration_days"`
	Price        int    `json:"price"`
}

func SubscriptionPlanToDTO(plan domain.SubscriptionPlan) SubscriptionPlanDTO {
	// Разделить Name на ServiceName и SubType по последнему пробелу
	name := plan.Name
	lastSpace := strings.LastIndex(name, " ")
	var serviceName, subType string
	if lastSpace > 0 {
		serviceName = name[:lastSpace]
		subType = name[lastSpace+1:]
	} else {
		serviceName = name
		subType = ""
	}
	return SubscriptionPlanDTO{
		PlanID:       int(plan.PlanID),
		ServiceID:    plan.ServiceID,
		Name:         name,
		ServiceName:  serviceName,
		SubType:      subType,
		DurationDays: plan.DurationDays,
		Price:        plan.Price,
	}
}

type PurchaseSubscriptionDTO struct {
	UserID       uuid.UUID `json:"user_id"`
	ServiceName  string    `json:"service_name"`
	PlanID       int       `json:"plan_id"`
	Price        int       `json:"price"`
	DurationDays int       `json:"duration_days"`
}

type PaymentDTO struct {
	PaymentID   int       `json:"payment_id"`
	UserID      uuid.UUID `json:"user_id"`
	CardNumber  string    `json:"card_number"`
	Amount      int       `json:"amount"`
	PaymentType string    `json:"payment_type"`
}
