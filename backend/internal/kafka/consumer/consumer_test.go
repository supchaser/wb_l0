package consumer

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"github.com/supchaser/wb_l0/internal/utils/pgxiface"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestCreateConsumer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.ConsumerConfig
		db      pgxiface.PgxIface
		wantErr bool
	}{
		{
			name:    "successful creation",
			cfg:     &config.ConsumerConfig{Brokers: []string{"localhost:9092"}, GroupID: "test-group", Topic: "test-topic", AutoOffsetReset: "earliest"},
			db:      nil,
			wantErr: false,
		},
		{
			name:    "nil config",
			cfg:     nil,
			db:      nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer, err := CreateConsumer(tt.cfg, tt.db)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConsumer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && consumer == nil {
				t.Error("CreateConsumer() returned nil consumer without error")
			}
		})
	}
}

func TestConsumer_ProcessSingleMessage(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mockDB.Close()

	consumer := &Consumer{
		db: mockDB,
	}

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	order := models.OrderRequest{
		OrderUID:          "test-order-123",
		TrackNumber:       "TRACK123",
		Entry:             "WBIL",
		Locale:            models.LocaleEN,
		InternalSignature: "internal_sig",
		CustomerID:        "test_customer",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		OofShard:          "1",
		DateCreated:       fixedTime,
		Delivery: models.DeliveryRequest{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "New York",
			Address: "Street 123",
			Region:  "NY",
			Email:   "john@example.com",
		},
		Payment: models.PaymentRequest{
			Transaction:  "trans-123",
			RequestID:    "req-123",
			Currency:     models.CurrencyUSD,
			Provider:     "wbpay",
			Amount:       1500,
			PaymentDt:    1,
			Bank:         "alpha",
			DeliveryCost: 500,
			GoodsTotal:   1000,
			CustomFee:    0,
		},
		Items: []models.ItemRequest{
			{
				ChrtID:      1,
				TrackNumber: "TRACK123",
				Price:       500,
				Rid:         "rid123",
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  500,
				NmID:        123456,
				Brand:       "Test Brand",
				Status:      202,
			},
		},
	}

	msgValue, _ := json.Marshal(order)
	msg := &kafka.Message{
		Value: msgValue,
		TopicPartition: kafka.TopicPartition{
			Topic:     stringPtr("test-topic"),
			Partition: 0,
			Offset:    123,
		},
	}

	mockDB.ExpectBegin()
	mockDB.ExpectQuery(`INSERT INTO "order"`).
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
			order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.OofShard, fixedTime,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mockDB.ExpectExec(`INSERT INTO delivery`).
		WithArgs(int64(1), order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
			order.Delivery.Address, order.Delivery.Region, order.Delivery.Email).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockDB.ExpectExec(`INSERT INTO payment`).
		WithArgs(int64(1), order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
			order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockDB.ExpectExec(`DELETE FROM item`).WithArgs(int64(1)).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mockDB.ExpectExec(`INSERT INTO item`).
		WithArgs(int64(1), order.Items[0].ChrtID, order.Items[0].TrackNumber, order.Items[0].Price, order.Items[0].Rid,
			order.Items[0].Name, order.Items[0].Sale, order.Items[0].Size, order.Items[0].TotalPrice,
			order.Items[0].NmID, order.Items[0].Brand, order.Items[0].Status).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockDB.ExpectCommit()

	messages := []*kafka.Message{msg}
	err = consumer.processMessageBatch(messages)
	if err != nil {
		t.Errorf("processMessageBatch() failed: %v", err)
	}

	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConsumer_ProcessSingleMessage_InvalidJSON(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mockDB.Close()

	consumer := &Consumer{
		db: mockDB,
	}

	msg := &kafka.Message{
		Value: []byte("invalid json"),
	}

	mockDB.ExpectBegin()

	ctx := context.Background()
	tx, err := mockDB.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	err = consumer.processSingleMessage(ctx, tx, msg)
	if err == nil {
		t.Error("expected error for invalid JSON, but got none")
	}

	tx.Rollback(ctx)

	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConsumer_ProcessSingleMessage_ValidationFailed(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mockDB.Close()

	consumer := &Consumer{
		db: mockDB,
	}

	invalidOrder := models.OrderRequest{
		OrderUID: "test-order",
	}
	msgValue, _ := json.Marshal(invalidOrder)
	msg := &kafka.Message{
		Value: msgValue,
	}

	mockDB.ExpectBegin()

	ctx := context.Background()
	tx, err := mockDB.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	err = consumer.processSingleMessage(ctx, tx, msg)
	if err == nil {
		t.Error("expected validation error, but got none")
	}

	tx.Rollback(ctx)

	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConsumer_ProcessMessageBatch(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mockDB.Close()

	consumer := &Consumer{
		db: mockDB,
	}

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	order1 := models.OrderRequest{
		OrderUID:          "order1",
		TrackNumber:       "TRACK1",
		Entry:             "WBIL",
		Locale:            models.LocaleEN,
		InternalSignature: "internal_sig1",
		CustomerID:        "cust1",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		OofShard:          "1",
		DateCreated:       fixedTime,
		Delivery: models.DeliveryRequest{
			Name:    "John",
			Phone:   "+79992222333",
			Zip:     "SDFSDF",
			City:    "SDFSDF",
			Address: "sdfgdfg",
			Region:  "fdsfgdfg",
			Email:   "s_sdfsdf@mail.ru",
		},
		Payment: models.PaymentRequest{
			Transaction:  "trans1",
			RequestID:    "req1",
			Currency:     models.CurrencyAUD,
			Provider:     "sdfsd",
			Amount:       1000,
			PaymentDt:    1234567890,
			Bank:         "bank1",
			DeliveryCost: 500,
			GoodsTotal:   1000,
			CustomFee:    0,
		},
		Items: []models.ItemRequest{{
			ChrtID:      1,
			TrackNumber: "TRACK1",
			Price:       500,
			Rid:         "rid1",
			Name:        "Item1",
			Sale:        0,
			Size:        "M",
			TotalPrice:  500,
			NmID:        123456,
			Brand:       "Brand1",
			Status:      202,
		}},
	}

	order2 := models.OrderRequest{
		OrderUID:          "order2",
		TrackNumber:       "TRACK2",
		Entry:             "WBIL",
		Locale:            models.LocaleEN,
		InternalSignature: "internal_sig2",
		CustomerID:        "cust2",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		OofShard:          "1",
		DateCreated:       fixedTime,
		Delivery: models.DeliveryRequest{
			Name:    "Jane",
			Phone:   "+79993333444",
			Zip:     "DFGHDF",
			City:    "DFGHDF",
			Address: "dfghdfg",
			Region:  "dfghdfg",
			Email:   "jane@mail.ru",
		},
		Payment: models.PaymentRequest{
			Transaction:  "trans2",
			RequestID:    "req2",
			Currency:     models.CurrencyUSD,
			Provider:     "provider2",
			Amount:       2000,
			PaymentDt:    1234567891,
			Bank:         "bank2",
			DeliveryCost: 300,
			GoodsTotal:   2000,
			CustomFee:    0,
		},
		Items: []models.ItemRequest{{
			ChrtID:      2,
			TrackNumber: "TRACK2",
			Price:       1000,
			Rid:         "rid2",
			Name:        "Item2",
			Sale:        0,
			Size:        "L",
			TotalPrice:  1000,
			NmID:        654321,
			Brand:       "Brand2",
			Status:      202,
		}},
	}

	msgValue1, _ := json.Marshal(order1)
	msgValue2, _ := json.Marshal(order2)

	topic := "test_topic"
	messages := []*kafka.Message{
		{
			Value: msgValue1,
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
				Offset:    1,
			},
		},
		{
			Value: msgValue2,
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
				Offset:    2,
			},
		},
	}

	mockDB.ExpectBegin()

	mockDB.ExpectQuery(`INSERT INTO "order"`).
		WithArgs(
			order1.OrderUID, order1.TrackNumber, order1.Entry, order1.Locale,
			order1.InternalSignature, order1.CustomerID, order1.DeliveryService,
			order1.Shardkey, order1.SmID, order1.OofShard, order1.DateCreated,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))

	mockDB.ExpectExec(`INSERT INTO delivery`).
		WithArgs(
			int64(1), order1.Delivery.Name, order1.Delivery.Phone, order1.Delivery.Zip,
			order1.Delivery.City, order1.Delivery.Address, order1.Delivery.Region, order1.Delivery.Email,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectExec(`INSERT INTO payment`).
		WithArgs(
			int64(1), order1.Payment.Transaction, order1.Payment.RequestID, order1.Payment.Currency,
			order1.Payment.Provider, order1.Payment.Amount, order1.Payment.PaymentDt, order1.Payment.Bank,
			order1.Payment.DeliveryCost, order1.Payment.GoodsTotal, order1.Payment.CustomFee,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectExec(`DELETE FROM item`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	mockDB.ExpectExec(`INSERT INTO item`).
		WithArgs(
			int64(1), order1.Items[0].ChrtID, order1.Items[0].TrackNumber, order1.Items[0].Price,
			order1.Items[0].Rid, order1.Items[0].Name, order1.Items[0].Sale, order1.Items[0].Size,
			order1.Items[0].TotalPrice, order1.Items[0].NmID, order1.Items[0].Brand, order1.Items[0].Status,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectQuery(`INSERT INTO "order"`).
		WithArgs(
			order2.OrderUID, order2.TrackNumber, order2.Entry, order2.Locale,
			order2.InternalSignature, order2.CustomerID, order2.DeliveryService,
			order2.Shardkey, order2.SmID, order2.OofShard, order2.DateCreated,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(2)))

	mockDB.ExpectExec(`INSERT INTO delivery`).
		WithArgs(
			int64(2), order2.Delivery.Name, order2.Delivery.Phone, order2.Delivery.Zip,
			order2.Delivery.City, order2.Delivery.Address, order2.Delivery.Region, order2.Delivery.Email,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectExec(`INSERT INTO payment`).
		WithArgs(
			int64(2), order2.Payment.Transaction, order2.Payment.RequestID, order2.Payment.Currency,
			order2.Payment.Provider, order2.Payment.Amount, order2.Payment.PaymentDt, order2.Payment.Bank,
			order2.Payment.DeliveryCost, order2.Payment.GoodsTotal, order2.Payment.CustomFee,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectExec(`DELETE FROM item`).
		WithArgs(int64(2)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	mockDB.ExpectExec(`INSERT INTO item`).
		WithArgs(
			int64(2), order2.Items[0].ChrtID, order2.Items[0].TrackNumber, order2.Items[0].Price,
			order2.Items[0].Rid, order2.Items[0].Name, order2.Items[0].Sale, order2.Items[0].Size,
			order2.Items[0].TotalPrice, order2.Items[0].NmID, order2.Items[0].Brand, order2.Items[0].Status,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectCommit()

	err = consumer.processMessageBatch(messages)
	if err != nil {
		t.Errorf("processMessageBatch() failed: %v", err)
	}

	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConsumer_Stop(t *testing.T) {
	mockConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "test-group",
	})
	if err != nil {
		t.Skip("Kafka not available, skipping test")
	}

	consumer := &Consumer{
		consumer: mockConsumer,
		stopChan: make(chan struct{}),
		wg:       sync.WaitGroup{},
	}

	consumer.wg.Add(1)
	go func() {
		defer consumer.wg.Done()
		<-consumer.stopChan
	}()

	consumer.Stop()
}

func stringPtr(s string) *string {
	return &s
}
