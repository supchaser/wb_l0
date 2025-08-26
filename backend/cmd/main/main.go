package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/supchaser/wb_l0/internal/app/delivery"
	"github.com/supchaser/wb_l0/internal/app/repository"
	"github.com/supchaser/wb_l0/internal/app/usecase"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/kafka/consumer"
	"github.com/supchaser/wb_l0/internal/middleware"
	"github.com/supchaser/wb_l0/internal/utils/db"
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

	dbpool, err := db.CreateConnectionPool(cfg)
	if err != nil {
		logger.Fatal("failed to connect to DB", zap.Error(err))
	}
	defer dbpool.Close()

	redisOpts, err := redis.ParseURL(cfg.RedisDSN)
	if err != nil {
		logger.Fatal("error connecting to Redis: ", zap.Error(err))
		return
	}
	redisDB := redis.NewClient(redisOpts)
	defer redisDB.Close()

	if err := redisDB.Ping(redisDB.Context()).Err(); err != nil {
		logger.Fatal("error while pinging Redis: ", zap.Error(err))
		return
	}

	kafkaConsumer, err := consumer.CreateConsumer(cfg.ConsumerConfig, dbpool)
	if err != nil {
		logger.Fatal("failed to create Kafka consumer", zap.Error(err))
	}

	go func() {
		if err := kafkaConsumer.Start(); err != nil {
			logger.Fatal("failed to start Kafka consumer", zap.Error(err))
		}
	}()

	appRepo := repository.CreateAppRepository(dbpool, redisDB)
	appUsecase := usecase.CreateAppUsecase(appRepo)
	appDelivery := delivery.CreateAppDelivery(appUsecase)

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	orderRouter := apiRouter.PathPrefix("/orders").Subrouter()
	orderRouter.HandleFunc("/{order_uid}", appDelivery.GetOrderByID).Methods("GET")

	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.PanicMiddleware)

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173", "http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: cors(router),
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting HTTP server",
			zap.String("address", "localhost"+server.Addr),
			zap.Any("config", cfg),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
			serverErr <- err
		}
	}()

	logger.Info("application started successfully",
		zap.String("http_port", cfg.ServerPort),
		zap.Strings("kafka_brokers", cfg.ConsumerConfig.Brokers),
		zap.String("kafka_topic", cfg.ConsumerConfig.Topic))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("failed to start server", zap.Error(err))
		os.Exit(1)
	case sig := <-quit:
		logger.Info("server is shutting down",
			zap.String("signal", sig.String()),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Info("stopping Kafka consumer...")
		kafkaConsumer.Stop()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
			return
		}

		logger.Info("server stopped")
		os.Exit(0)
	}
}
