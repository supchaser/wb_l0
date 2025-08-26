package config

import (
	"os"
	"testing"
)

func TestCheckEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVars   []string
		setup     func()
		cleanup   func()
		wantError bool
	}{
		{
			name:      "all env vars present",
			envVars:   []string{"TEST_VAR_1", "TEST_VAR_2"},
			setup:     func() { os.Setenv("TEST_VAR_1", "value1"); os.Setenv("TEST_VAR_2", "value2") },
			cleanup:   func() { os.Unsetenv("TEST_VAR_1"); os.Unsetenv("TEST_VAR_2") },
			wantError: false,
		},
		{
			name:      "missing env var",
			envVars:   []string{"TEST_VAR_1", "MISSING_VAR"},
			setup:     func() { os.Setenv("TEST_VAR_1", "value1") },
			cleanup:   func() { os.Unsetenv("TEST_VAR_1") },
			wantError: true,
		},
		{
			name:      "empty env var value",
			envVars:   []string{"TEST_VAR_1", "EMPTY_VAR"},
			setup:     func() { os.Setenv("TEST_VAR_1", "value1"); os.Setenv("EMPTY_VAR", "") },
			cleanup:   func() { os.Unsetenv("TEST_VAR_1"); os.Unsetenv("EMPTY_VAR") },
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			err := checkEnv(tt.envVars)
			if (err != nil) != tt.wantError {
				t.Errorf("checkEnv() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateEnv(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		cleanup   func()
		wantError bool
	}{
		{
			name: "all required env vars present",
			setup: func() {
				os.Setenv("LOG_MODE", "dev")
				os.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/db")
				os.Setenv("REDIS_DSN", "redis://localhost:6379")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
			},
			cleanup: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("POSTGRES_DSN")
				os.Unsetenv("REDIS_DSN")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
			},
			wantError: false,
		},
		{
			name: "missing required env var",
			setup: func() {
				os.Setenv("LOG_MODE", "dev")
				os.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/db")
				// Missing REDIS_DSN
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
			},
			cleanup: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("POSTGRES_DSN")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			err := validateEnv()
			if (err != nil) != tt.wantError {
				t.Errorf("validateEnv() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		setup        func()
		cleanup      func()
		want         string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			setup:        func() { os.Setenv("TEST_VAR", "actual") },
			cleanup:      func() { os.Unsetenv("TEST_VAR") },
			want:         "actual",
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			setup:        func() {},
			cleanup:      func() {},
			want:         "default",
		},
		{
			name:         "env var empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setup:        func() { os.Setenv("EMPTY_VAR", "") },
			cleanup:      func() { os.Unsetenv("EMPTY_VAR") },
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			if got := getEnv(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		setup        func()
		cleanup      func()
		want         int
	}{
		{
			name:         "valid int env var",
			key:          "INT_VAR",
			defaultValue: 100,
			setup:        func() { os.Setenv("INT_VAR", "42") },
			cleanup:      func() { os.Unsetenv("INT_VAR") },
			want:         42,
		},
		{
			name:         "invalid int env var",
			key:          "INVALID_INT_VAR",
			defaultValue: 100,
			setup:        func() { os.Setenv("INVALID_INT_VAR", "not_a_number") },
			cleanup:      func() { os.Unsetenv("INVALID_INT_VAR") },
			want:         100,
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_INT_VAR",
			defaultValue: 100,
			setup:        func() {},
			cleanup:      func() {},
			want:         100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			if got := getEnvInt(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnvInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		setup        func()
		cleanup      func()
		want         bool
	}{
		{
			name:         "true bool env var",
			key:          "BOOL_VAR",
			defaultValue: false,
			setup:        func() { os.Setenv("BOOL_VAR", "true") },
			cleanup:      func() { os.Unsetenv("BOOL_VAR") },
			want:         true,
		},
		{
			name:         "false bool env var",
			key:          "BOOL_VAR",
			defaultValue: true,
			setup:        func() { os.Setenv("BOOL_VAR", "false") },
			cleanup:      func() { os.Unsetenv("BOOL_VAR") },
			want:         false,
		},
		{
			name:         "invalid bool env var",
			key:          "INVALID_BOOL_VAR",
			defaultValue: true,
			setup:        func() { os.Setenv("INVALID_BOOL_VAR", "not_a_bool") },
			cleanup:      func() { os.Unsetenv("INVALID_BOOL_VAR") },
			want:         true,
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_BOOL_VAR",
			defaultValue: true,
			setup:        func() {},
			cleanup:      func() {},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			if got := getEnvBool(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		cleanup   func()
		wantError bool
		validate  func(*testing.T, *Config)
	}{
		{
			name: "successful config load",
			setup: func() {
				os.Setenv("LOG_MODE", "test")
				os.Setenv("POSTGRES_DSN", "postgres://test:test@localhost:5432/test")
				os.Setenv("REDIS_DSN", "redis://localhost:6379/0")
				os.Setenv("SERVER_PORT", "9090")
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka1:9092,kafka2:9092")
				os.Setenv("PRODUCER_PORT", "9091")
				os.Setenv("KAFKA_CLIENT_ID", "test-client")
				os.Setenv("KAFKA_ACKS", "1")
				os.Setenv("KAFKA_COMPRESSION_TYPE", "gzip")
				os.Setenv("KAFKA_RETRIES", "5")
				os.Setenv("KAFKA_BATCH_SIZE", "32768")
				os.Setenv("KAFKA_LINGER_MS", "200")
				os.Setenv("KAFKA_ENABLE_IDEMPOTENCE", "false")
				os.Setenv("TOPIC", "test-topic")
				os.Setenv("KAFKA_CONSUMER_GROUP_ID", "test-group")
				os.Setenv("KAFKA_AUTO_OFFSET_RESET", "latest")
				os.Setenv("KAFKA_ENABLE_AUTO_COMMIT", "true")
			},
			cleanup: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("POSTGRES_DSN")
				os.Unsetenv("REDIS_DSN")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
				os.Unsetenv("PRODUCER_PORT")
				os.Unsetenv("KAFKA_CLIENT_ID")
				os.Unsetenv("KAFKA_ACKS")
				os.Unsetenv("KAFKA_COMPRESSION_TYPE")
				os.Unsetenv("KAFKA_RETRIES")
				os.Unsetenv("KAFKA_BATCH_SIZE")
				os.Unsetenv("KAFKA_LINGER_MS")
				os.Unsetenv("KAFKA_ENABLE_IDEMPOTENCE")
				os.Unsetenv("TOPIC")
				os.Unsetenv("KAFKA_CONSUMER_GROUP_ID")
				os.Unsetenv("KAFKA_AUTO_OFFSET_RESET")
				os.Unsetenv("KAFKA_ENABLE_AUTO_COMMIT")
			},
			wantError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.LogMode != "test" {
					t.Errorf("LogMode = %v, want %v", cfg.LogMode, "test")
				}
				if cfg.ServerPort != "9090" {
					t.Errorf("ServerPort = %v, want %v", cfg.ServerPort, "9090")
				}
				if cfg.ProducerPort != "9091" {
					t.Errorf("ProducerPort = %v, want %v", cfg.ProducerPort, "9091")
				}
				if len(cfg.ProducerConfig.Brokers) != 2 {
					t.Errorf("ProducerConfig.Brokers length = %v, want %v", len(cfg.ProducerConfig.Brokers), 2)
				}
				if cfg.ProducerConfig.ClientID != "test-client" {
					t.Errorf("ProducerConfig.ClientID = %v, want %v", cfg.ProducerConfig.ClientID, "test-client")
				}
				if cfg.ConsumerConfig.GroupID != "test-group" {
					t.Errorf("ConsumerConfig.GroupID = %v, want %v", cfg.ConsumerConfig.GroupID, "test-group")
				}
			},
		},
		{
			name: "missing required env vars",
			setup: func() {
				os.Setenv("LOG_MODE", "test")
				os.Setenv("POSTGRES_DSN", "postgres://test:test@localhost:5432/test")
			},
			cleanup: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("POSTGRES_DSN")
			},
			wantError: true,
			validate:  nil,
		},
		{
			name: "default values used",
			setup: func() {
				os.Setenv("LOG_MODE", "dev")
				os.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/db")
				os.Setenv("REDIS_DSN", "redis://localhost:6379")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
			},
			cleanup: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("POSTGRES_DSN")
				os.Unsetenv("REDIS_DSN")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
			},
			wantError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ProducerConfig.ClientID != "wb-l0-producer" {
					t.Errorf("ProducerConfig.ClientID = %v, want %v", cfg.ProducerConfig.ClientID, "wb-l0-producer")
				}
				if cfg.ProducerConfig.Acks != "all" {
					t.Errorf("ProducerConfig.Acks = %v, want %v", cfg.ProducerConfig.Acks, "all")
				}
				if cfg.ProducerConfig.Retries != 3 {
					t.Errorf("ProducerConfig.Retries = %v, want %v", cfg.ProducerConfig.Retries, 3)
				}
				if cfg.ConsumerConfig.GroupID != "wb-l0-consumer-group" {
					t.Errorf("ConsumerConfig.GroupID = %v, want %v", cfg.ConsumerConfig.GroupID, "wb-l0-consumer-group")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			cfg, err := LoadConfig()
			if (err != nil) != tt.wantError {
				t.Errorf("LoadConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadConfig_KafkaBrokersParsing(t *testing.T) {
	setup := func() {
		os.Setenv("LOG_MODE", "dev")
		os.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("REDIS_DSN", "redis://localhost:6379")
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "broker1:9092,broker2:9092,broker3:9092")
	}

	cleanup := func() {
		os.Unsetenv("LOG_MODE")
		os.Unsetenv("POSTGRES_DSN")
		os.Unsetenv("REDIS_DSN")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
	}

	setup()
	defer cleanup()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if len(cfg.ProducerConfig.Brokers) != 3 {
		t.Errorf("Expected 3 brokers, got %d", len(cfg.ProducerConfig.Brokers))
	}

	if len(cfg.ConsumerConfig.Brokers) != 3 {
		t.Errorf("Expected 3 brokers, got %d", len(cfg.ConsumerConfig.Brokers))
	}

	expectedBrokers := []string{"broker1:9092", "broker2:9092", "broker3:9092"}
	for i, broker := range expectedBrokers {
		if cfg.ProducerConfig.Brokers[i] != broker {
			t.Errorf("Producer broker[%d] = %v, want %v", i, cfg.ProducerConfig.Brokers[i], broker)
		}
		if cfg.ConsumerConfig.Brokers[i] != broker {
			t.Errorf("Consumer broker[%d] = %v, want %v", i, cfg.ConsumerConfig.Brokers[i], broker)
		}
	}
}
