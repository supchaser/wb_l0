package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestGetOrderByID_FromCache(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	orderUID := "test-order-123"
	expectedOrder := &models.Order{
		OrderUID:    orderUID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
	}

	orderJSON, err := json.Marshal(expectedOrder)
	assert.NoError(t, err)

	redisMock.ExpectGet(fmt.Sprintf("order:%s", orderUID)).SetVal(string(orderJSON))

	ctx := context.Background()
	result, err := repo.GetOrderByID(ctx, orderUID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderUID, result.OrderUID)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetOrderByID_FromDB_NotFound(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	orderUID := "non-existent-order"

	redisMock.ExpectGet(fmt.Sprintf("order:%s", orderUID)).SetErr(redis.Nil)

	pgxMock.ExpectBegin()

	orderQuery := `SELECT id, order_uid, track_number, entry, locale, internal_signature,
			   customer_id, delivery_service, shardkey, sm_id, oof_shard,
			   date_created, updated_at
		FROM "order" 
		WHERE order_uid = \$1`
	pgxMock.ExpectQuery(orderQuery).WithArgs(orderUID).WillReturnError(pgx.ErrNoRows)

	pgxMock.ExpectRollback()

	ctx := context.Background()
	result, err := repo.GetOrderByID(ctx, orderUID)

	assert.ErrorIs(t, err, errs.ErrNotFound)
	assert.Nil(t, result)
	assert.NoError(t, redisMock.ExpectationsWereMet())
	assert.NoError(t, pgxMock.ExpectationsWereMet())
}

func TestGetOrderByID_FromDB_Success(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	orderUID := "test-order-123"
	now := time.Now()

	redisMock.ExpectGet(fmt.Sprintf("order:%s", orderUID)).SetErr(redis.Nil)

	pgxMock.ExpectBegin()

	orderQuery := `SELECT id, order_uid, track_number, entry, locale, internal_signature,
			   customer_id, delivery_service, shardkey, sm_id, oof_shard,
			   date_created, updated_at
		FROM "order" 
		WHERE order_uid = \$1`
	orderRows := pgxmock.NewRows([]string{
		"id", "order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "oof_shard",
		"date_created", "updated_at",
	}).AddRow(
		int64(1), orderUID, "WBILMTESTTRACK", "WBIL", "en", "test_signature",
		"test_customer", "test_service", "test_shard", int64(123), "test_oof",
		now, now,
	)
	pgxMock.ExpectQuery(orderQuery).
		WithArgs(orderUID).
		WillReturnRows(orderRows)

	deliveryQuery := `SELECT id, name, phone, zip, city, address, region, email,
			   created_at, updated_at
		FROM delivery 
		WHERE order_id = \$1`
	pgxMock.ExpectQuery(deliveryQuery).
		WithArgs(int64(1)).
		WillReturnError(pgx.ErrNoRows)

	paymentQuery := `SELECT id, transaction, request_id, currency, provider, amount,
			   payment_dt, bank, delivery_cost, goods_total, custom_fee,
			   created_at, updated_at
		FROM payment 
		WHERE order_id = \$1`
	pgxMock.ExpectQuery(paymentQuery).
		WithArgs(int64(1)).
		WillReturnError(pgx.ErrNoRows)

	itemsQuery := `SELECT id, chrt_id, track_number, price, rid, name, sale, size,
			   total_price, nm_id, brand, status, created_at, updated_at
		FROM item 
		WHERE order_id = \$1
		ORDER BY id`
	itemsRows := pgxmock.NewRows([]string{
		"id", "chrt_id", "track_number", "price", "rid", "name", "sale", "size",
		"total_price", "nm_id", "brand", "status", "created_at", "updated_at",
	})
	pgxMock.ExpectQuery(itemsQuery).
		WithArgs(int64(1)).
		WillReturnRows(itemsRows)

	pgxMock.ExpectCommit()

	redisMock.Regexp().ExpectSet(fmt.Sprintf("order:%s", orderUID), `.*`, 7*24*time.Hour).SetVal("OK")

	ctx := context.Background()
	result, err := repo.GetOrderByID(ctx, orderUID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderUID, result.OrderUID)
	assert.NoError(t, redisMock.ExpectationsWereMet())
	assert.NoError(t, pgxMock.ExpectationsWereMet())
}

func TestGetOrderByID_RedisError(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	orderUID := "test-order-123"
	redisMock.ExpectGet(fmt.Sprintf("order:%s", orderUID)).SetErr(errors.New("redis connection error"))

	ctx := context.Background()
	result, err := repo.GetOrderByID(ctx, orderUID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetOrderFromCache_InvalidJSON(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	orderUID := "test-order-123"
	redisMock.ExpectGet(fmt.Sprintf("order:%s", orderUID)).SetVal("invalid json")
	redisMock.ExpectDel(fmt.Sprintf("order:%s", orderUID)).SetVal(1)

	ctx := context.Background()
	result, err := repo.getOrderFromCache(ctx, orderUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestSaveOrderToCache_Success(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, redisClient)

	order := &models.Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
	}

	orderJSON, err := json.Marshal(order)
	assert.NoError(t, err)

	redisMock.ExpectSet(fmt.Sprintf("order:%s", order.OrderUID), orderJSON, 7*24*time.Hour).SetVal("OK")

	ctx := context.Background()
	err = repo.saveOrderToCache(ctx, order)

	assert.NoError(t, err)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetOrderFromDB_DeliveryError(t *testing.T) {
	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgx mock: %v", err)
	}
	defer pgxMock.Close(context.Background())

	repo := CreateAppRepository(pgxMock, nil)

	orderUID := "test-order-123"
	now := time.Now()

	pgxMock.ExpectBegin()

	orderQuery := `SELECT id, order_uid, track_number, entry, locale, internal_signature,
           customer_id, delivery_service, shardkey, sm_id, oof_shard,
           date_created, updated_at
    FROM "order" 
    WHERE order_uid = \$1`
	orderRows := pgxmock.NewRows([]string{
		"id", "order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "oof_shard",
		"date_created", "updated_at",
	}).AddRow(
		int64(1),
		orderUID, "WBILMTESTTRACK", "WBIL", "en", "test_signature",
		"test_customer", "test_service", "test_shard", int64(123), "test_oof",
		now, now,
	)
	pgxMock.ExpectQuery(orderQuery).
		WithArgs(orderUID).
		WillReturnRows(orderRows)

	deliveryQuery := `SELECT id, name, phone, zip, city, address, region, email,
           created_at, updated_at
    FROM delivery 
    WHERE order_id = \$1`
	pgxMock.ExpectQuery(deliveryQuery).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	pgxMock.ExpectRollback()

	ctx := context.Background()
	result, err := repo.getOrderFromDB(ctx, orderUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, pgxMock.ExpectationsWereMet())
}
