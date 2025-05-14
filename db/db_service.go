package db

import (
	"context"

	"github.com/eric-schulze/we_sync_bricks/config"
	"github.com/eric-schulze/we_sync_bricks/pgdb"
)

type DBService struct {
	DBClient DBClient
}

type DBClient interface {
	StartConnectionPool() (string, error)
}

func NewDBService(dbConfig config.DBConfig, context *context.Context) (*DBService, error) {
	dbClient, err := pgdb.NewPostgresDbClient(config.DBConfig, context)
	return &DBService{
		DBClient: dbClient,
	}, nil
}