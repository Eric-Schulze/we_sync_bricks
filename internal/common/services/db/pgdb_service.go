package db

import (
	"context"
	"log"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGDBService struct {
	pgxClient *pgxpool.Pool
	context context.Context
	Config models.DBConfig
}

func NewPostgresDbClient(dbConfig *models.DBConfig, context *context.Context) (*PGDBService, error) {
	db, err := NewPg(*context, dbConfig, WithPgxConfig(dbConfig))
	if err != nil {
		log.Fatal("Error while creating connection to the database!!")
	}

	return &PGDBService{
		pgxClient: db,
		context: *context,
		Config: *dbConfig,
	}, nil
}

// StartConnectionPool creates a new pool of connections to the Postgres db
// with the defined number of connections and settings
// TODO Refactor to wrap the pgxpool.Conn in custom type so that 
// db_service.go does not have to depend on pgxpool
func (client PGDBService) StartConnectionPool() (*pgxpool.Conn, error) {
	connection, err := client.pgxClient.Acquire(context.Background())
	if err!=nil {
	 log.Fatal("Error while acquiring connection from the database pool!!")
	} 
	defer connection.Release()
   
	err = connection.Ping(context.Background())
	if err!=nil{
	 log.Fatal("Could not ping database")

	}

	return connection, nil
}

// TODO Refactor to wrap the pgx.Rows in custom type so that 
// db_service.go does not have to depend on pgx
func (client PGDBService) Query(query string, args ...any) (pgx.Rows, error) {
	rows, err := client.pgxClient.Query(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return nil, err
	}
	// Note: rows.Close() should be called by the caller, not here

	return rows, nil
}

// TODO Refactor to wrap the pgx.Rows in custom type so that 
// db_service.go does not have to depend on pgx
func (client PGDBService) ExecSQL(query string, args ...any) (string, error) {
	status, err := client.pgxClient.Exec(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return status.String(), err
	}

	return status.String(), nil
}

// CollectRowsToStruct executes a query and collects the results into a slice of structs
// using pgx.RowToStructByName. This encapsulates the pgx.CollectRows functionality.
// Usage: results, err := db.CollectRowsToStruct[MyStruct](query, args...)
func CollectRowsToStruct[T any](client PGDBService, query string, args ...any) ([]T, error) {
	rows, err := client.pgxClient.Query(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		logger.Error("Error collecting rows to struct", "query", query, "error", err)
		return nil, err
	}

	return results, nil
}

// CollectRowsToStructByPos executes a query and collects the results into a slice of structs
// using pgx.RowToStructByPos. This is useful when you want to map by position instead of name.
// Usage: results, err := db.CollectRowsToStructByPos[MyStruct](dbService, query, args...)
func CollectRowsToStructByPos[T any](client PGDBService, query string, args ...any) ([]T, error) {
	rows, err := client.pgxClient.Query(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByPos[T])
	if err != nil {
		logger.Error("Error collecting rows to struct by position", "query", query, "error", err)
		return nil, err
	}

	return results, nil
}

// CollectRowsToMap executes a query and collects the results into a slice of maps
func (client PGDBService) CollectRowsToMap(query string, args ...any) ([]map[string]any, error) {
	rows, err := client.pgxClient.Query(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		logger.Error("Error collecting rows to map", "query", query, "error", err)
		return nil, err
	}

	return results, nil
}

// QueryRowToStruct executes a query that returns a single row and maps it to a struct
// Usage: result, err := db.QueryRowToStruct[MyStruct](dbService, query, args...)
func QueryRowToStruct[T any](client PGDBService, query string, args ...any) (T, error) {
	var result T
	rows, err := client.pgxClient.Query(client.context, query, args...)
	if err != nil {
		logger.Error("Error in PGDB query", "query", query, "error", err)
		return result, err
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Error("No rows found", "query", query)
		return result, pgx.ErrNoRows
	}

	result, err = pgx.RowToStructByName[T](rows)
	if err != nil {
		logger.Error("Error scanning row to struct", "query", query, "error", err)
		return result, err
	}

	return result, nil
}