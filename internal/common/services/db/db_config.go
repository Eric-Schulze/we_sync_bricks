package db

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure singleton pattern for connection configuration
var pgOnce sync.Once

func WithPgxConfig(dbConfig *models.DBConfig) *pgx.ConnConfig {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	
	// Create the dsn string
	dbConfig.DSN = strings.TrimSpace(fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d",
		dbConfig.UserName, dbConfig.Password, dbConfig.DBName,
		dbConfig.Host, dbConfig.Port))

	pgxConfig, err := pgx.ParseConfig(dbConfig.DSN)
	if err!=nil {
	 	logger.Error("Failed to create a config", "error", err)
		return nil
	}
   
	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifeTime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
   
	return pgxConfig
}

// Create a new connection pool with the provided configuration
func NewPg(ctx context.Context, dbConfig *models.DBConfig, pgxConfig *pgx.ConnConfig) (*pgxpool.Pool, error) {
	// Parse the pool configuration from connection string
	config, err := pgxpool.ParseConfig(pgxConfig.ConnString())
	if err != nil {
		logger.Error("Error parsing pool config", "error", err)
		return nil, err
	}

	// Apply pool-specific configurations
	config.MaxConns = dbConfig.MaxConns
	config.MinConns = dbConfig.MinConns
	config.MaxConnLifetime = dbConfig.MaxConnLifeTime
	config.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	config.HealthCheckPeriod = dbConfig.HealthCheckPeriod
   
	config.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
	 	logger.Info("Before acquiring the connection pool to the database!!")
		return true
	}
   
	config.AfterRelease = func(c *pgx.Conn) bool {
	 	logger.Info("After releasing the connection pool to the database!!")
	 	return true
	}
   
	config.BeforeClose = func(c *pgx.Conn) {
		logger.Info("Closed the connection pool to the database!!")
	}

	logger.Info("config", "config", config)

	// Initialize the pool with singleton pattern
	var db *pgxpool.Pool
	pgOnce.Do(func() {
		db, err = pgxpool.NewWithConfig(ctx, config)
	})

	// Verify the connection
	if err = db.Ping(ctx); err != nil {
		logger.Error("Unable to ping database", "error", err)
		return nil, err
	}
	logger.Info("Successfully connected to database")

	return db, nil
}