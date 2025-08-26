package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/supchaser/wb_l0/internal/config"
	"github.com/supchaser/wb_l0/internal/utils/logger"

	"go.uber.org/zap"
)

func CreateConnectionPool(config *config.Config) (*pgxpool.Pool, error) {
	postgresDSN := config.PostgresDSN
	dbpool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		logger.Fatalf("failed to connect DB with DSN: ", postgresDSN, "error is", zap.Error(err))
	}

	logger.Info("successful connection to the database")

	return dbpool, nil
}
