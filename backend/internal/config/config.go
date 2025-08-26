package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LogMode               string
	ServerPort            string
	ProducerPort          string
	PostgresDSN           string
	RedisDSN              string
	KafkaBootstrapServers string

	ProducerConfig *ProducerConfig
	ConsumerConfig *ConsumerConfig
}

type ProducerConfig struct {
	Brokers           []string
	ClientID          string
	Acks              string
	CompressionType   string
	Topic             string
	Retries           int
	BatchSize         int
	LingerMs          int
	EnableIdempotence bool
}

type ConsumerConfig struct {
	Brokers          []string
	GroupID          string
	Topic            string
	AutoOffsetReset  string
	EnableAutoCommit bool
}

func checkEnv(envVars []string) error {
	var missingVars []string

	for _, envVar := range envVars {
		if value, exists := os.LookupEnv(envVar); !exists || value == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("error: this env vars are missing: %v", missingVars)
	}
	return nil
}

func validateEnv() error {
	err := checkEnv([]string{
		"LOG_MODE",
		"POSTGRES_DSN",
		"REDIS_DSN",
		"SERVER_PORT",
		"KAFKA_BOOTSTRAP_SERVERS",
	})
	if err != nil {
		return err
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func LoadConfig() (*Config, error) {
	err := validateEnv()
	if err != nil {
		return nil, fmt.Errorf("LoadConfig: %w", err)
	}

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"), ",")

	return &Config{
		LogMode:               getEnv("LOG_MODE", "dev"),
		ServerPort:            getEnv("SERVER_PORT", "8080"),
		ProducerPort:          getEnv("PRODUCER_PORT", "8081"),
		PostgresDSN:           os.Getenv("POSTGRES_DSN"),
		RedisDSN:              os.Getenv("REDIS_DSN"),
		KafkaBootstrapServers: os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),

		ProducerConfig: &ProducerConfig{
			Brokers:           kafkaBrokers,
			ClientID:          getEnv("KAFKA_CLIENT_ID", "wb-l0-producer"),
			Acks:              getEnv("KAFKA_ACKS", "all"),
			CompressionType:   getEnv("KAFKA_COMPRESSION_TYPE", "snappy"),
			Retries:           getEnvInt("KAFKA_RETRIES", 3),
			BatchSize:         getEnvInt("KAFKA_BATCH_SIZE", 16384),
			LingerMs:          getEnvInt("KAFKA_LINGER_MS", 100),
			EnableIdempotence: getEnvBool("KAFKA_ENABLE_IDEMPOTENCE", true),
			Topic:             getEnv("TOPIC", "orders"),
		},

		ConsumerConfig: &ConsumerConfig{
			Brokers:          kafkaBrokers,
			GroupID:          getEnv("KAFKA_CONSUMER_GROUP_ID", "wb-l0-consumer-group"),
			Topic:            getEnv("KAFKA_TOPIC", "orders"),
			AutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
			EnableAutoCommit: getEnvBool("KAFKA_ENABLE_AUTO_COMMIT", false),
		},
	}, nil
}
