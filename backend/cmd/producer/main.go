package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/kafka/producer"
	"github.com/supchaser/wb_l0/internal/utils/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("error initializing config: %v\n", err)
		os.Exit(1)
	}

	err = logger.Init(cfg.LogMode)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("configuration loaded successfully")
	logger.Debug("debug mode enabled",
		zap.String("log_mode", cfg.LogMode),
		zap.String("server_port", cfg.ServerPort),
	)

	producer, err := producer.CreateProducer(cfg.ProducerConfig)
	if err != nil {
		logger.Fatal("failed to create producer", zap.Error(err))
	}
	defer producer.Close()

	if err := producer.HealthCheck(context.Background()); err != nil {
		logger.Fatal("health check failed", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go generateOrders(ctx, producer)

	<-sigChan
	logger.Info("shutting down producer...")
}

func generateOrders(ctx context.Context, producer *producer.Producer) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			order := generateTestOrder(i, localRand)
			if err := producer.Produce(ctx, order, producer.Config.Topic); err != nil {
				logger.Error("failed to produce order",
					zap.Error(err),
					zap.String("order_id", order.OrderUID))
			} else {
				logger.Info("order produced successfully",
					zap.String("order_id", order.OrderUID),
					zap.String("track_number", order.TrackNumber))
			}
		}
	}
}

func generateTestOrder(seq int, r *rand.Rand) models.OrderRequest {
	orderUID := fmt.Sprintf("b563feb7b2b84b6test%d", seq)
	trackNumber := fmt.Sprintf("WBILMTESTTRACK%d", seq)
	currentTime := time.Now()

	itemsCount := r.Intn(3) + 1
	items := make([]models.ItemRequest, itemsCount)

	for i := range itemsCount {
		items[i] = generateTestItem(trackNumber, i, r)
	}

	return models.OrderRequest{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Locale:            getRandomLocale(r),
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer%d", r.Intn(1000)),
		DeliveryService:   getRandomDeliveryService(r),
		Shardkey:          fmt.Sprintf("%d", r.Intn(10)),
		SmID:              r.Intn(100),
		DateCreated:       currentTime,
		OofShard:          fmt.Sprintf("%d", r.Intn(2)),
		Delivery:          generateTestDelivery(r),
		Payment:           generateTestPayment(orderUID, r),
		Items:             items,
	}
}

func generateTestDelivery(r *rand.Rand) models.DeliveryRequest {
	firstNames := []string{"John", "Jane", "Alex", "Maria", "David", "Sarah", "Mike", "Anna"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
	cities := []string{"Moscow", "Saint Petersburg", "Novosibirsk", "Yekaterinburg", "Kazan"}
	regions := []string{"Moscow Oblast", "Leningrad Oblast", "Sverdlovsk Oblast", "Republic of Tatarstan"}

	return models.DeliveryRequest{
		Name:    fmt.Sprintf("%s %s", firstNames[r.Intn(len(firstNames))], lastNames[r.Intn(len(lastNames))]),
		Phone:   fmt.Sprintf("+7%d", 9000000000+r.Int63n(100000000)),
		Zip:     fmt.Sprintf("%d", 100000+r.Intn(900000)),
		City:    cities[r.Intn(len(cities))],
		Address: fmt.Sprintf("%s st., %d", getRandomStreet(r), r.Intn(100)+1),
		Region:  regions[r.Intn(len(regions))],
		Email:   fmt.Sprintf("test%d@example.com", r.Intn(1000)),
	}
}

func generateTestPayment(orderUID string, r *rand.Rand) models.PaymentRequest {
	amount := r.Intn(10000) + 1000

	return models.PaymentRequest{
		Transaction:  orderUID,
		RequestID:    "",
		Currency:     getRandomCurrency(r),
		Provider:     "wbpay",
		Amount:       amount,
		PaymentDt:    int(time.Now().Unix()),
		Bank:         getRandomBank(r),
		DeliveryCost: r.Intn(500) + 100,
		GoodsTotal:   amount - 500,
		CustomFee:    0,
	}
}

func generateTestItem(trackNumber string, index int, r *rand.Rand) models.ItemRequest {
	products := []struct {
		name  string
		brand string
		price int
	}{
		{"Smartphone", "Samsung", 25000},
		{"Laptop", "Apple", 150000},
		{"Headphones", "Sony", 5000},
		{"Watch", "Casio", 3000},
		{"Camera", "Canon", 45000},
		{"Tablet", "Huawei", 20000},
		{"Speaker", "JBL", 7000},
	}

	product := products[r.Intn(len(products))]

	return models.ItemRequest{
		ChrtID:      r.Intn(10000000),
		TrackNumber: trackNumber,
		Price:       product.price,
		Rid:         fmt.Sprintf("ab4219087a764ae0btest%d", index),
		Name:        product.name,
		Sale:        r.Intn(30),
		Size:        fmt.Sprintf("%d", r.Intn(5)),
		TotalPrice:  product.price - (product.price * r.Intn(30) / 100),
		NmID:        r.Intn(1000000),
		Brand:       product.brand,
		Status:      202,
	}
}

func getRandomLocale(r *rand.Rand) models.LocaleEnum {
	locales := []models.LocaleEnum{
		models.LocaleEN,
		models.LocaleRU,
		models.LocaleES,
		models.LocaleDE,
		models.LocaleFR,
	}
	return locales[r.Intn(len(locales))]
}

func getRandomCurrency(r *rand.Rand) models.CurrencyEnum {
	currencies := []models.CurrencyEnum{
		models.CurrencyUSD,
		models.CurrencyEUR,
		models.CurrencyRUB,
		models.CurrencyGBP,
	}
	return currencies[r.Intn(len(currencies))]
}

func getRandomDeliveryService(r *rand.Rand) string {
	services := []string{"meest", "russian-post", "dhl", "fedex", "ups", "cdek"}
	return services[r.Intn(len(services))]
}

func getRandomBank(r *rand.Rand) string {
	banks := []string{"sberbank", "alpha", "tinkoff", "vtb", "gazprombank", "raiffeisen"}
	return banks[r.Intn(len(banks))]
}

func getRandomStreet(r *rand.Rand) string {
	streets := []string{"Lenin", "Pushkin", "Gorky", "Peace", "Victory", "Central", "Green", "Sunset"}
	return streets[r.Intn(len(streets))]
}
