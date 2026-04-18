package domain

import (
	"context"
	"time"
)

type StatsService interface {
	GetServiceStatsFromDB(ctx context.Context) (*StatsResponse, error)
}

type TopService struct {
	ServiceName string
	SubCount    int
	Revenue     int
}

type ServiceStats struct {
	ServiceName  string
	TotalSubs    int
	ActiveSubs   int
	TotalRevenue int
}

type StatsResponse struct {
	TopServices   []TopService
	ServiceStats  []ServiceStats
	TotalRevenue  int
	TotalSubs     int
	GeneratedAt   time.Time
	ExecutionTime time.Duration
	Source        string // "db" or "cache"
}
