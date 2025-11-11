package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manish-npx/go-echo-pg/internal/config"
)

func ConnectPostgres(cfg *config.Config) *pgxpool.Pool {
	d := cfg.DB
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}

	// set pool config using values from config
	if cfg.DBMaxIdleConns > 0 {
		pool.Config().MaxConns = int32(cfg.DBMaxIdleConns)
	}
	// You can configure other pool parameters as needed.

	// simple ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	log.Println("✅ connected to postgres")
	return pool
}
