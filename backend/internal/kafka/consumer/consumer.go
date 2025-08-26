package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"github.com/supchaser/wb_l0/internal/utils/pgxiface"
	"github.com/supchaser/wb_l0/internal/utils/validate"
	"go.uber.org/zap"
)

const (
	sessionTimeout     = 6000
	autoCommitInterval = 1000
	pollTimeout        = 100
	batchSize          = 1000
)

type Consumer struct {
	consumer  *kafka.Consumer
	config    *config.ConsumerConfig
	db        pgxiface.PgxIface
	wg        sync.WaitGroup
	stopChan  chan struct{}
	batchChan chan *kafka.Message
}

func CreateConsumer(cfg *config.ConsumerConfig, db pgxiface.PgxIface) (*Consumer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("consumer config is required")
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":         strings.Join(cfg.Brokers, ","),
		"group.id":                  cfg.GroupID,
		"session.timeout.ms":        sessionTimeout,
		"auto.offset.reset":         cfg.AutoOffsetReset,
		"enable.auto.commit":        cfg.EnableAutoCommit,
		"auto.commit.interval.ms":   autoCommitInterval,
		"max.poll.interval.ms":      300000,
		"heartbeat.interval.ms":     3000,
		"max.partition.fetch.bytes": 1048576,
		"fetch.message.max.bytes":   10485760,
	}

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	consumer := &Consumer{
		consumer:  c,
		config:    cfg,
		db:        db,
		stopChan:  make(chan struct{}),
		batchChan: make(chan *kafka.Message, batchSize),
	}

	return consumer, nil
}

func (c *Consumer) Start() error {
	topics := []string{c.config.Topic}
	if err := c.consumer.SubscribeTopics(topics, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	logger.Info("starting Kafka consumer",
		zap.String("group_id", c.config.GroupID),
		zap.Strings("topics", topics),
		zap.Strings("brokers", c.config.Brokers))

	c.wg.Add(1)
	go c.batchProcessor()

	c.wg.Add(1)
	go c.messageLoop()

	logger.Info("Kafka consumer started successfully")
	return nil
}

func (c *Consumer) messageLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			logger.Info("stopping message loop")
			return

		default:
			msg, err := c.consumer.ReadMessage(pollTimeout)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				logger.Error("failed to read message", zap.Error(err))
				continue
			}

			select {
			case c.batchChan <- msg:
			default:
				logger.Warn("batch channel full, message might be processed slowly")
				c.batchChan <- msg
			}
		}
	}
}

func (c *Consumer) batchProcessor() {
	defer c.wg.Done()

	batch := make([]*kafka.Message, 0, batchSize)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	processBatch := func() {
		if len(batch) == 0 {
			return
		}

		if err := c.processMessageBatch(batch); err != nil {
			logger.Error("failed to process batch", zap.Error(err))
		}

		if err := c.commitOffsets(batch); err != nil {
			logger.Error("failed to commit offsets", zap.Error(err))
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-c.stopChan:
			processBatch()
			logger.Info("batch processor stopped")
			return

		case msg := <-c.batchChan:
			batch = append(batch, msg)
			if len(batch) >= batchSize {
				processBatch()
			}

		case <-ticker.C:
			processBatch()
		}
	}
}

func (c *Consumer) processMessageBatch(messages []*kafka.Message) error {
	ctx := context.Background()
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if err := c.processSingleMessage(ctx, tx, msg); err != nil {
			var topic string
			if msg.TopicPartition.Topic != nil {
				topic = *msg.TopicPartition.Topic
			}
			logger.Error("failed to process message",
				zap.String("topic", topic),
				zap.Int32("partition", msg.TopicPartition.Partition),
				zap.Int64("offset", int64(msg.TopicPartition.Offset)),
				zap.Error(err))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("successfully processed message batch",
		zap.Int("message_count", len(messages)))

	return nil
}

func (c *Consumer) processSingleMessage(ctx context.Context, tx pgx.Tx, msg *kafka.Message) error {
	var order models.OrderRequest
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	if err := validate.ValidateOrderRequest(&order); err != nil {
		logger.Warn("order validation failed",
			zap.String("order_uid", order.OrderUID),
			zap.Error(err))
		return fmt.Errorf("order validation failed: %w", err)
	}

	if err := c.saveOrderToDB(ctx, tx, &order); err != nil {
		return fmt.Errorf("failed to save order to DB: %w", err)
	}

	logger.Info("successfully processed order",
		zap.String("order_uid", order.OrderUID),
		zap.String("topic", *msg.TopicPartition.Topic),
		zap.Int32("partition", msg.TopicPartition.Partition),
		zap.Int64("offset", int64(msg.TopicPartition.Offset)))

	return nil
}

func (c *Consumer) saveOrderToDB(ctx context.Context, tx pgx.Tx, order *models.OrderRequest) error {
	orderID, err := c.saveMainOrder(ctx, tx, order)
	if err != nil {
		return fmt.Errorf("failed to save main order: %w", err)
	}

	if err := c.saveDelivery(ctx, tx, orderID, order.Delivery); err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	if err := c.savePayment(ctx, tx, orderID, order.Payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	if err := c.saveItems(ctx, tx, orderID, order.Items); err != nil {
		return fmt.Errorf("failed to save items: %w", err)
	}

	return nil
}

func (c *Consumer) saveMainOrder(ctx context.Context, tx pgx.Tx, order *models.OrderRequest) (int64, error) {
	query := `
        INSERT INTO "order" (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, oof_shard, date_created
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            oof_shard = EXCLUDED.oof_shard,
            date_created = EXCLUDED.date_created,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `

	var orderID int64
	err := tx.QueryRow(ctx, query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.OofShard,
		order.DateCreated,
	).Scan(&orderID)

	if err != nil {
		return 0, fmt.Errorf("failed to insert/update order: %w", err)
	}

	return orderID, nil
}

func (c *Consumer) saveDelivery(ctx context.Context, tx pgx.Tx, orderID int64, delivery models.DeliveryRequest) error {
	query := `
        INSERT INTO delivery (
            order_id, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_id) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email,
            updated_at = CURRENT_TIMESTAMP
    `

	_, err := tx.Exec(ctx, query,
		orderID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)

	if err != nil {
		return fmt.Errorf("failed to insert/update delivery: %w", err)
	}

	return nil
}

func (c *Consumer) savePayment(ctx context.Context, tx pgx.Tx, orderID int64, payment models.PaymentRequest) error {
	query := `
        INSERT INTO payment (
            order_id, transaction, request_id, currency, provider,
            amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (transaction) DO UPDATE SET
            order_id = EXCLUDED.order_id,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee,
            updated_at = CURRENT_TIMESTAMP
    `

	_, err := tx.Exec(ctx, query,
		orderID,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)

	if err != nil {
		return fmt.Errorf("failed to insert/update payment: %w", err)
	}

	return nil
}

func (c *Consumer) saveItems(ctx context.Context, tx pgx.Tx, orderID int64, items []models.ItemRequest) error {
	deleteQuery := `DELETE FROM item WHERE order_id = $1`
	_, err := tx.Exec(ctx, deleteQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	query := `
        INSERT INTO item (
            order_id, chrt_id, track_number, price, rid, name,
            sale, size, total_price, nm_id, brand, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

	for _, item := range items {
		_, err := tx.Exec(ctx, query,
			orderID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	return nil
}

func (c *Consumer) commitOffsets(messages []*kafka.Message) error {
	if c.config.EnableAutoCommit {
		return nil
	}

	offsets := make([]kafka.TopicPartition, 0, len(messages))
	for _, msg := range messages {
		offset := kafka.TopicPartition{
			Topic:     msg.TopicPartition.Topic,
			Partition: msg.TopicPartition.Partition,
			Offset:    msg.TopicPartition.Offset + 1,
		}
		offsets = append(offsets, offset)
	}

	_, err := c.consumer.CommitOffsets(offsets)
	return err
}

func (c *Consumer) Stop() {
	close(c.stopChan)
	c.consumer.Close()
	c.wg.Wait()
	logger.Info("Kafka consumer stopped")
}

func (c *Consumer) HealthCheck() error {
	metadata, err := c.consumer.GetMetadata(nil, true, 5000)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	if len(metadata.Topics) == 0 {
		return fmt.Errorf("no topics available")
	}

	return nil
}
