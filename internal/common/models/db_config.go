package models

import (
	"time"
)

// DBConfig holds all database configuration parameters
type DBConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	UserName          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	DBName            string `mapstructure:"dbname"`
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifeTime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}
