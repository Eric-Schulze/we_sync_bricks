package config

import (
	"context"
	"log"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBConfig holds all database configuration parameters
type DBConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	UserName          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	DBName            string `mapstructure:"dbname"`
	MaxConns          int32
	MinConns          int32
	MaxConnLifeTime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func NewDBConfig() (*DBConfig, error) {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5
	
    // Your own Database URL
	const DATABASE_URL string = "postgres://postgres:12345678@localhost:5432/postgres?"
   
	dbConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err!=nil {
	 	log.Fatal("Failed to create a config, error: ", err)
	}
   
	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout
   
	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
	 	log.Println("Before acquiring the connection pool to the database!!")
		return true
	}
   
	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
	 	log.Println("After releasing the connection pool to the database!!")
	 	return true
	}
   
	dbConfig.BeforeClose = func(c *pgx.Conn) {
		log.Println("Closed the connection pool to the database!!")
	}
   
	return dbConfig
}


// Ensure singleton pattern for connection configuration
var pgOnce sync.Once

// Create a pgx connection config from DBConfig
func WithPgxConfig(dbConfig *DBConfig) *pgx.ConnConfig {
	// Create the dsn string
	connString := strings.TrimSpace(fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d",
		dbConfig.UserName, dbConfig.Password, dbConfig.DBName,
		dbConfig.Host, dbConfig.Port))

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		log.Error("Error parsing connection config", slog.String("error", err.Error()))
		panic(err)
	}

	return config
}

// Create a new connection pool with the provided configuration
func NewPg(ctx context.Context, dbConfig *DBConfig, pgxConfig *pgx.ConnConfig) (*pgxpool.Pool, error) {
	// Parse the pool configuration from connection string
	config, err := pgxpool.ParseConfig(pgxConfig.ConnString())
	if err != nil {
		log.Error("Error parsing pool config", slog.String("error", err.Error()))
		return nil, err
	}

	// Apply pool-specific configurations
	config.MaxConns = dbConfig.MaxConns
	config.MinConns = dbConfig.MinConns
	config.MaxConnLifetime = dbConfig.MaxConnLifeTime
	config.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	config.HealthCheckPeriod = dbConfig.HealthCheckPeriod

	// Initialize the pool with singleton pattern
	var db *pgxpool.Pool
	pgOnce.Do(func() {
		db, err = pgxpool.NewWithConfig(ctx, config)
	})

	// Verify the connection
	if err = db.Ping(ctx); err != nil {
		log.Error("Unable to ping database", log.String("error", err.Error()))
		return nil, err
	}
	log.Info("Successfully connected to database")

	return db, nil
}