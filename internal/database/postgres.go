package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"go.uber.org/zap"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(config *config.Config, logger *zap.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host,
		config.DB.Port,
		config.DB.User,
		config.DB.Password,
		config.DB.DBName,
		config.DB.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(config.DB.MaxConns)
	poolConfig.MinConns = int32(config.DB.MinConns)

	if config.DB.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = config.DB.MaxConnLifetime
	}
	if config.DB.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = config.DB.MaxConnIdleTime
	}

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", config.DB.Host),
		zap.String("database", config.DB.DBName),
		zap.Int("max_connections", config.DB.MaxConns),
	)

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// HealthCheck checks if database is responsive
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
