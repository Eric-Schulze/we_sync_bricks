package pgdb

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eric-schulze/we_sync_bricks/config"
)

func NewPostgresDbClient(dbConfig config.DBConfig) *pgxpool.Pool {
	dbPool, err := pgxpool.ConnectConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatal("Error while creating connection to the database!!")
	}
	return dbPool
}

// StartConnectionPool creates a new pool of connections to the Postgres db
// with the defined number of connections and settings
func NewConnection() (*pgxpool.Conn, error) {
	// Create database connection
	connPool,err := pgxpool.NewWithConfig(context.Background(), Config())
	if err!=nil {
	 log.Fatal("Error while creating connection to the database!!")
	} 
   
	connection, err := connPool.Acquire(context.Background())
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
