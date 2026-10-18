package domain

import "context"

// Service — онлайн-сервис, на который оформляется подписка.
type Service struct {
	ServiceID     int
	ServiceName   string
	UsersCount    int
	HasPromocodes bool
}

// ServiceRepository — репозиторий сервисов (таблица services).
type ServiceRepository interface {
	StoreService(ctx context.Context, s Service) (int, error)
	GetService(ctx context.Context, serviceID int) (Service, error)
	GetServiceByName(ctx context.Context, name string) (Service, error)
	ListServices(ctx context.Context) ([]Service, error)
	UpdateService(ctx context.Context, s Service) error
	DeleteService(ctx context.Context, serviceID int) error
}
