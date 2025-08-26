package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/supchaser/wb_l0/internal/app"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"github.com/supchaser/wb_l0/internal/utils/validate"
	"go.uber.org/zap"
)

type AppUsecase struct {
	orderRepository app.AppRepository
}

func CreateAppUsecase(orderRepository app.AppRepository) *AppUsecase {
	return &AppUsecase{
		orderRepository: orderRepository,
	}
}

func (uc *AppUsecase) GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error) {
	const funcName = "Usecase.GetOrderByID"

	if err := validate.ValidateOrderUID(orderUID); err != nil {
		logger.Warn("invalid order UID",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		return nil, fmt.Errorf("%w: %v", errs.ErrValidation, err)
	}

	order, err := uc.orderRepository.GetOrderByID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			logger.Warn("order not found",
				zap.String("function", funcName),
				zap.String("order_uid", orderUID))
			return nil, errs.ErrNotFound
		}

		logger.Error("failed to get order",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		return nil, fmt.Errorf("%s: failed to get order: %w", funcName, err)
	}

	logger.Info("order retrieved successfully",
		zap.String("function", funcName),
		zap.String("order_uid", orderUID))

	return order, nil
}
