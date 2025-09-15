package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
)

type SubscriptionDTO struct {
	SubId       int
	UserId      uuid.UUID
	ServiceName string
	Price       int
	SubType     string
	StartDate   time.Time
	EndDate     time.Time
}

type SubsFilterDTO struct {
	StartDate   time.Time
	EndDate     time.Time
	UserID      uuid.UUID
	ServiceName string
	SubType     string
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
