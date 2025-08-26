package delivery

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/supchaser/wb_l0/internal/app"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"github.com/supchaser/wb_l0/internal/utils/responses"
	"go.uber.org/zap"
)

type AppDelivery struct {
	orderUsecase app.AppUsecase
}

func CreateAppDelivery(orderUsecase app.AppUsecase) *AppDelivery {
	return &AppDelivery{
		orderUsecase: orderUsecase,
	}
}

func (d *AppDelivery) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	const funcName = "AppDelivery.GetOrderByID"

	logger.Info("handling get order request",
		zap.String("function", funcName),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))

	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	if orderUID == "" {
		responses.DoBadResponseAndLog(w, http.StatusBadRequest, "order_uid is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	order, err := d.orderUsecase.GetOrderByID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			responses.DoBadResponseAndLog(w, http.StatusNotFound, "order not found")
			return
		}

		logger.Error("failed to get order",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		responses.DoBadResponseAndLog(w, http.StatusInternalServerError, "internal server error")
		return
	}

	orderResponse := d.convertToResponse(order)

	responses.DoJSONResponse(w, orderResponse, http.StatusOK)

	logger.Info("order retrieved successfully",
		zap.String("function", funcName),
		zap.String("order_uid", orderUID))
}

func (d *AppDelivery) convertToResponse(order *models.Order) map[string]any {
	return map[string]any{
		"order_uid":          order.OrderUID,
		"track_number":       order.TrackNumber,
		"entry":              order.Entry,
		"locale":             order.Locale,
		"internal_signature": order.InternalSignature,
		"customer_id":        order.CustomerID,
		"delivery_service":   order.DeliveryService,
		"shardkey":           order.Shardkey,
		"sm_id":              order.SmID,
		"date_created":       order.DateCreated.Format(time.RFC3339),
		"oof_shard":          order.OofShard,
		"delivery":           d.convertDeliveryToResponse(order.Delivery),
		"payment":            d.convertPaymentToResponse(order.Payment),
		"items":              d.convertItemsToResponse(order.Items),
	}
}

func (d *AppDelivery) convertDeliveryToResponse(delivery *models.Delivery) map[string]any {
	if delivery == nil {
		return nil
	}

	return map[string]any{
		"name":    delivery.Name,
		"phone":   delivery.Phone,
		"zip":     delivery.Zip,
		"city":    delivery.City,
		"address": delivery.Address,
		"region":  delivery.Region,
		"email":   delivery.Email,
	}
}

func (d *AppDelivery) convertPaymentToResponse(payment *models.Payment) map[string]any {
	if payment == nil {
		return nil
	}

	return map[string]any{
		"transaction":   payment.Transaction,
		"request_id":    payment.RequestID,
		"currency":      payment.Currency,
		"provider":      payment.Provider,
		"amount":        payment.Amount,
		"payment_dt":    payment.PaymentDt,
		"bank":          payment.Bank,
		"delivery_cost": payment.DeliveryCost,
		"goods_total":   payment.GoodsTotal,
		"custom_fee":    payment.CustomFee,
	}
}

func (d *AppDelivery) convertItemsToResponse(items []models.Item) []map[string]any {
	if items == nil {
		return nil
	}

	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"chrt_id":      item.ChrtID,
			"track_number": item.TrackNumber,
			"price":        item.Price,
			"rid":          item.Rid,
			"name":         item.Name,
			"sale":         item.Sale,
			"size":         item.Size,
			"total_price":  item.TotalPrice,
			"nm_id":        item.NmID,
			"brand":        item.Brand,
			"status":       item.Status,
		})
	}

	return result
}
