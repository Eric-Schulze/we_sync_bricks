package models


import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBService interface {
	StartConnectionPool() (*pgxpool.Conn, error)
	Query(query string, args ...any) (pgx.Rows, error)
	ExecSQL(string, ...any) (string, error)
	// CollectRows wrapper methods
	CollectRowsToMap(query string, args ...any) ([]map[string]any, error)
}