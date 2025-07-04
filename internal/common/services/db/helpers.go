package db

import (
	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/jackc/pgx/v5"
)

// CollectRowsToStructFromService is a helper function that works with the DBService interface
// It abstracts away the need to cast to concrete types in repositories
func CollectRowsToStructFromService[T any](dbService models.DBService, query string, args ...any) ([]T, error) {
	// Check if we can cast to our concrete implementation to use the optimized method
	if pgdbService, ok := dbService.(*PGDBService); ok {
		return CollectRowsToStruct[T](*pgdbService, query, args...)
	}

	// Fallback: use the interface method and convert manually
	rows, err := dbService.Query(query)
	if err != nil {
		logger.Error("Error in database query", "query", query, "error", err)
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

// QueryRowToStructFromService is a helper function for single row queries that works with the DBService interface
func QueryRowToStructFromService[T any](dbService models.DBService, query string, args ...any) (T, error) {
	var result T
	
	// Check if we can cast to our concrete implementation to use the optimized method
	if pgdbService, ok := dbService.(*PGDBService); ok {
		return QueryRowToStruct[T](*pgdbService, query, args...)
	}

	// Fallback: use the interface method to get the first row
	rows, err := dbService.Query(query, args...)
	if err != nil {
		logger.Error("Error in database query", "query", query, "error", err)
		return result, err
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Error("No rows found", "query", query)
		return result, pgx.ErrNoRows
	}

	// Use pgx.RowToStructByName to scan into the struct
	result, err = pgx.RowToStructByName[T](rows)
	if err != nil {
		logger.Error("Error scanning row to struct", "query", query, "error", err)
		return result, err
	}

	return result, nil
}