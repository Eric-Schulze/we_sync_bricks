package db

import (
	"context"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
)

func NewDBService(dbConfig *models.DBConfig, context *context.Context) (models.DBService, error) {
	dbService, err := NewPostgresDbClient(dbConfig, context)
	if err != nil {
		return nil, err
	}

	return dbService, nil
}
