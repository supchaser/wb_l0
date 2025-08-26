package app

import (
	"context"

	"github.com/supchaser/wb_l0/internal/app/models"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go

type AppRepository interface {
	GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error)
}

type AppUsecase interface {
	GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error)
}
