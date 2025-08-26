package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"github.com/supchaser/wb_l0/internal/utils/pgxiface"
	"go.uber.org/zap"
)

type AppRepository struct {
	postgresDB pgxiface.PgxIface
	redisDB    *redis.Client
}

func CreateAppRepository(postgresDB pgxiface.PgxIface, redisDB *redis.Client) *AppRepository {
	return &AppRepository{
		postgresDB: postgresDB,
		redisDB:    redisDB,
	}
}

func (ar *AppRepository) GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error) {
	const funcName = "GetOrderByID"

	if order, err := ar.getOrderFromCache(ctx, orderUID); err == nil {
		logger.Info("order found in cache",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID))
		return order, nil
	}

	order, err := ar.getOrderFromDB(ctx, orderUID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			logger.Warn("order not found",
				zap.String("function", funcName),
				zap.String("order_uid", orderUID))
			return nil, errs.ErrNotFound
		}
		logger.Error("failed to get order from database",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		return nil, fmt.Errorf("%s: failed to get order: %w", funcName, err)
	}

	if err := ar.saveOrderToCache(ctx, order); err != nil {
		logger.Warn("failed to save order to cache",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
	}

	logger.Info("order retrieved from database and cached",
		zap.String("function", funcName),
		zap.String("order_uid", orderUID))

	return order, nil
}

func (ar *AppRepository) getOrderFromCache(ctx context.Context, orderUID string) (*models.Order, error) {
	const funcName = "getOrderFromCache"

	cacheKey := fmt.Sprintf("order:%s", orderUID)

	data, err := ar.redisDB.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errs.ErrNotFound
		}
		logger.Warn("redis get error",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		return nil, fmt.Errorf("%s: redis error: %w", funcName, err)
	}

	order := models.Order{}
	if err := json.Unmarshal(data, &order); err != nil {
		logger.Warn("failed to unmarshal order from cache",
			zap.String("function", funcName),
			zap.String("order_uid", orderUID),
			zap.Error(err))
		ar.redisDB.Del(ctx, cacheKey)
		return nil, fmt.Errorf("%s: unmarshal error: %w", funcName, err)
	}

	return &order, nil
}

func (ar *AppRepository) getOrderFromDB(ctx context.Context, orderUID string) (*models.Order, error) {
	const funcName = "getOrderFromDB"

	tx, err := ar.postgresDB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", funcName, err)
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		SELECT id, order_uid, track_number, entry, locale, internal_signature,
			   customer_id, delivery_service, shardkey, sm_id, oof_shard,
			   date_created, updated_at
		FROM "order" 
		WHERE order_uid = $1
	`

	order := &models.Order{}
	err = tx.QueryRow(ctx, orderQuery, orderUID).Scan(
		&order.ID,
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.OofShard,
		&order.DateCreated,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("%s: failed to get order: %w", funcName, err)
	}

	deliveryQuery := `
		SELECT id, name, phone, zip, city, address, region, email,
			   created_at, updated_at
		FROM delivery 
		WHERE order_id = $1
	`

	delivery := &models.Delivery{}
	err = tx.QueryRow(ctx, deliveryQuery, order.ID).Scan(
		&delivery.ID,
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: failed to get delivery: %w", funcName, err)
	}
	order.Delivery = delivery

	paymentQuery := `
		SELECT id, transaction, request_id, currency, provider, amount,
			   payment_dt, bank, delivery_cost, goods_total, custom_fee,
			   created_at, updated_at
		FROM payment 
		WHERE order_id = $1
	`

	payment := &models.Payment{}
	err = tx.QueryRow(ctx, paymentQuery, order.ID).Scan(
		&payment.ID,
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDt,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: failed to get payment: %w", funcName, err)
	}
	order.Payment = payment

	itemsQuery := `
		SELECT id, chrt_id, track_number, price, rid, name, sale, size,
			   total_price, nm_id, brand, status, created_at, updated_at
		FROM item 
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := tx.Query(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get items: %w", funcName, err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan item: %w", funcName, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", funcName, err)
	}
	order.Items = items

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", funcName, err)
	}

	return order, nil
}

func (ar *AppRepository) saveOrderToCache(ctx context.Context, order *models.Order) error {
	const funcName = "saveOrderToCache"

	cacheKey := fmt.Sprintf("order:%s", order.OrderUID)

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal order: %w", funcName, err)
	}

	err = ar.redisDB.Set(ctx, cacheKey, data, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("%s: failed to save to redis: %w", funcName, err)
	}

	logger.Debug("order saved to cache",
		zap.String("function", funcName),
		zap.String("order_uid", order.OrderUID))

	return nil
}
