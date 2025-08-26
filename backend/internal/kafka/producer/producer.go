package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"go.uber.org/zap"
)

const (
	defaultFlushTimeout    = 5000
	defaultDeliveryTimeout = 10 * time.Second
	maxRetries             = 3
)

type Producer struct {
	producer     *kafka.Producer
	Config       *config.ProducerConfig
	wg           sync.WaitGroup
	closeOnce    sync.Once
	deliveryChan chan kafka.Event
}

func CreateProducer(cfg *config.ProducerConfig) (*Producer, error) {
	if cfg == nil {
		cfg = &config.ProducerConfig{}
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":      strings.Join(cfg.Brokers, ","),
		"client.id":              cfg.ClientID,
		"acks":                   cfg.Acks,
		"retries":                cfg.Retries,
		"enable.idempotence":     cfg.EnableIdempotence,
		"compression.type":       cfg.CompressionType,
		"batch.size":             cfg.BatchSize,
		"linger.ms":              cfg.LingerMs,
		"message.timeout.ms":     int(defaultDeliveryTimeout / time.Millisecond),
		"go.delivery.reports":    true,
		"go.logs.channel.enable": true,
	}

	if _, err := kafkaConfig.Get("acks", nil); err != nil {
		if err := kafkaConfig.SetKey("acks", "all"); err != nil {
			return nil, fmt.Errorf("failed to set acks: %w", err)
		}
		logger.Info("set default acks to 'all'")
	}

	if _, err := kafkaConfig.Get("enable.idempotence", nil); err != nil {
		if err := kafkaConfig.SetKey("enable.idempotence", true); err != nil {
			return nil, fmt.Errorf("failed to set idempotence: %w", err)
		}
		logger.Info("set default idempotence to true")
	}

	if _, err := kafkaConfig.Get("retries", nil); err != nil {
		if err := kafkaConfig.SetKey("retries", maxRetries); err != nil {
			return nil, fmt.Errorf("failed to set retries: %w", err)
		}
		logger.Info("set default retries", zap.Int("retries", maxRetries))
	}

	p, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		logger.Error("failed to create Kafka producer", zap.Error(err))
		return nil, fmt.Errorf("error to create producer: %w", err)
	}

	producer := &Producer{
		producer:     p,
		Config:       cfg,
		deliveryChan: make(chan kafka.Event, 1000),
	}

	producer.startDeliveryHandler()

	logger.Info("kafka producer created successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("client_id", cfg.ClientID))

	return producer, nil
}

func (p *Producer) Produce(ctx context.Context, order models.OrderRequest, topic string) error {
	logger.Debug("producing order to Kafka",
		zap.String("order_uid", order.OrderUID),
		zap.String("topic", topic))

	orderInBytes, err := json.Marshal(order)
	if err != nil {
		logger.Error("failed to marshal order",
			zap.String("order_uid", order.OrderUID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: orderInBytes,
		Key:   []byte(order.OrderUID),
		Headers: []kafka.Header{
			{Key: "version", Value: []byte("1.0")},
			{Key: "content-type", Value: []byte("application/json")},
		},
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDeliveryTimeout)
	defer cancel()

	return p.produceWithRetry(ctx, message, maxRetries)
}

func (p *Producer) produceWithRetry(ctx context.Context, message *kafka.Message, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			logger.Warn("produce operation cancelled by context")
			return errs.ErrContextTimeout

		default:
			deliveryChan := make(chan kafka.Event, 1)

			if err := p.producer.Produce(message, deliveryChan); err != nil {
				lastErr = fmt.Errorf("produce failed on attempt %d: %w", attempt, err)
				logger.Warn("produce attempt failed",
					zap.Int("attempt", attempt),
					zap.Int("max_retries", maxRetries),
					zap.Error(err))

				delay := time.Duration(attempt)*100*time.Millisecond + time.Duration(rand.Int63n(50))*time.Millisecond

				logger.Debug("waiting before retry",
					zap.Duration("delay", delay),
					zap.Int("attempt", attempt))

				time.Sleep(delay)
				continue
			}

			select {
			case <-ctx.Done():
				logger.Warn("produce operation cancelled during delivery wait")
				return errs.ErrContextTimeout

			case ev := <-deliveryChan:
				switch e := ev.(type) {
				case *kafka.Message:
					logger.Info("message successfully delivered to Kafka",
						zap.String("topic", *e.TopicPartition.Topic),
						zap.Int32("partition", e.TopicPartition.Partition),
						zap.Int64("offset", int64(e.TopicPartition.Offset)),
						zap.String("key", string(message.Key)))
					return nil

				case kafka.Error:
					lastErr = e
					logger.Warn("kafka delivery error",
						zap.Int("attempt", attempt),
						zap.Bool("retriable", e.IsRetriable()),
						zap.Error(e))

					delay := time.Duration(attempt)*100*time.Millisecond + time.Duration(rand.Int63n(50))*time.Millisecond

					logger.Debug("waiting before retry after delivery error",
						zap.Duration("delay", delay),
						zap.Int("attempt", attempt))

					time.Sleep(delay)

				default:
					lastErr = fmt.Errorf("unexpected event type: %T", e)
					logger.Error("unexpected event type during delivery",
						zap.String("event_type", fmt.Sprintf("%T", e)),
						zap.Any("event", e))
				}
			}
		}
	}

	logger.Error("failed to produce message after all retries",
		zap.Int("max_retries", maxRetries),
		zap.Error(lastErr))

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (p *Producer) startDeliveryHandler() {
	p.wg.Go(func() {
		defer p.wg.Done()
		logger.Info("starting Kafka delivery event handler")

		for e := range p.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.Error("message delivery failed",
						zap.String("topic", *ev.TopicPartition.Topic),
						zap.Int32("partition", ev.TopicPartition.Partition),
						zap.Error(ev.TopicPartition.Error),
						zap.String("key", string(ev.Key)))
				} else {
					logger.Debug("message delivery confirmed",
						zap.String("topic", *ev.TopicPartition.Topic),
						zap.Int32("partition", ev.TopicPartition.Partition),
						zap.Int64("offset", int64(ev.TopicPartition.Offset)),
						zap.String("key", string(ev.Key)))
				}
			case kafka.Error:
				if ev.IsFatal() {
					logger.Error("fatal Kafka error", zap.Error(ev))
				} else {
					logger.Warn("kafka error", zap.Error(ev))
				}
			default:
				logger.Debug("received Kafka event",
					zap.String("event_type", fmt.Sprintf("%T", ev)),
					zap.Any("event", ev))
			}
		}

		logger.Info("kafka delivery event handler stopped")
	})
}

func (p *Producer) HealthCheck(ctx context.Context) error {
	logger.Debug("performing Kafka producer health check")

	metadata, err := p.producer.GetMetadata(nil, true, 5000)
	if err != nil {
		logger.Error("kafka producer health check failed", zap.Error(err))
		return fmt.Errorf("health check failed: %w", err)
	}

	logger.Info("kafka producer health check successful",
		zap.Int("brokers_count", len(metadata.Brokers)),
		zap.Int("topics_count", len(metadata.Topics)))

	return nil
}

func (p *Producer) Close() {
	p.closeOnce.Do(func() {
		logger.Info("shutting down Kafka producer...")

		remaining := p.producer.Flush(defaultFlushTimeout)
		if remaining > 0 {
			logger.Warn("messages remained in queue after flush",
				zap.Int("remaining_messages", remaining))
		} else {
			logger.Debug("all messages flushed successfully")
		}

		p.producer.Close()
		close(p.deliveryChan)
		p.wg.Wait()
		logger.Info("kafka producer shutdown complete")
	})
}

func (p *Producer) BatchProduce(ctx context.Context, orders []models.OrderRequest, topic string) error {
	logger.Info("starting batch produce operation",
		zap.Int("orders_count", len(orders)),
		zap.String("topic", topic))

	startTime := time.Now()
	wg := sync.WaitGroup{}
	errChan := make(chan error, len(orders))
	semaphore := make(chan struct{}, 10)
	successCount := 0
	failCount := 0

	for i, order := range orders {
		wg.Add(1)
		go func(order models.OrderRequest, index int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := p.Produce(ctx, order, topic); err != nil {
				logger.Error("failed to produce order in batch",
					zap.String("order_uid", order.OrderUID),
					zap.Int("index", index),
					zap.Error(err))

				select {
				case errChan <- err:
				default:
				}
				failCount++
			} else {
				successCount++
				if successCount%10 == 0 {
					logger.Debug("batch progress",
						zap.Int("success", successCount),
						zap.Int("failed", failCount),
						zap.Int("total", len(orders)))
				}
			}
		}(order, i)
	}

	wg.Wait()
	close(errChan)

	duration := time.Since(startTime)

	if len(errChan) > 0 {
		logger.Error("batch produce completed with errors",
			zap.Int("success_count", successCount),
			zap.Int("fail_count", failCount),
			zap.Int("total", len(orders)),
			zap.Duration("duration", duration))
		return <-errChan
	}

	logger.Info("batch produce completed successfully",
		zap.Int("orders_count", len(orders)),
		zap.Duration("duration", duration),
		zap.Float64("orders_per_second", float64(len(orders))/duration.Seconds()))

	return nil
}
