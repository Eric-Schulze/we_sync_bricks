package db

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/spf13/viper"
)

var createDatabaseMigrationName = "0001_InitialMigration.sql"

type migration struct {
	file_name  string
	version    string
	name       string
	created_at string
}

// loadDBConfigForMigration loads database configuration from config file
// This is a copy of the logic from internal/init/config.go to avoid circular dependencies
func loadDBConfigForMigration() (models.DBConfig, error) {
	// Set file name and type for environment configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./internal/init")

	viper.AutomaticEnv()

	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Handle errors reading the config file
		logger.Error("Fatal error config file", "error", err)
		return models.DBConfig{}, err
	}

	var dbConfig models.DBConfig
	if err := viper.UnmarshalKey("db", &dbConfig); err != nil {
		return models.DBConfig{}, err
	}

	return dbConfig, nil
}

// RunMigrate runs all pending migrations
func RunMigrate(args []string) {
	// Create a root context with cancellation
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load database configuration
	dbConfig, err := loadDBConfigForMigration()
	if err != nil {
		logger.Error("Error loading database configuration", "error", err)
		return
	}

	db, err := NewPostgresDbClient(&dbConfig, &rootCtx)
	if err != nil {
		logger.Error("Error creating database client", "error", err)
		return
	}

	migrateDB(*db)
}

// migrateDB finds the last run migration, and run all those after it in order
// We use the migrations table to do this
func migrateDB(db PGDBService) {
	var completed []string

	// Get a list of migration files
	files, err := filepath.Glob("./internal/common/services/db/migrations/*.sql")
	if err != nil {
		logger.Error("Error running finding migration files", "error", err)
		return
	}

	// Sort files alphabetically
	sort.Strings(files)

	// Check if this is a blank database (no migrations table)
	if !migrationsTableExists(db) {
		logger.Info("Database is blank - running all migrations from the beginning")
		// Run all migrations
	} else {
		// Get existing migrations and find where to start
		var migrations []string
		for _, migration := range getMigrations(db) {
			migrations = append(migrations, migration.file_name)
		}

		if len(migrations) > 0 {
			// Sort migrations alphabetically
			sort.Strings(migrations)

			// Find the index of the last migration and start from the next one
			lastMigration := migrations[len(migrations)-1]
			startIndex := slices.Index(files, lastMigration)
			if startIndex >= 0 && startIndex < len(files)-1 {
				files = files[startIndex+1:]
			} else {
				// All migrations are already run
				files = []string{}
			}
		}
	}

	for _, file := range files {
		filename := filepath.Base(file)

		logger.Info("Running migration", "file", filename)

		args := []string{"-d", db.Config.DSN, "-f", file}
		if strings.Contains(filename, createDatabaseMigrationName) {
			args = []string{"-f", file}
			logger.Info("Running database creation migration", "creation_file", file)
		}

		// Execute this sql file against the database
		result, err := runCommand("psql", args...)
		if err != nil || strings.Contains(string(result), "ERROR") {
			if err == nil {
				err = fmt.Errorf("%s", string(result))
			}

			// If at any point we fail, log it and break
			logger.Error("ERROR loading sql migration", "error", err)
			logger.Error("All further migrations cancelled")
			break
		}

		completed = append(completed, filename)
		logger.Info("Completed migration", "migration_file", filename, "result", string(result))
	}

	if len(completed) > 0 {
		updateMigrations(db, completed)
		logger.Info("Migrations complete up to migration " + completed[len(completed)-1] + " on db")
	} else {
		logger.Info("No migrations to perform")
	}
}

// migrationsTableExists checks if the migrations table exists in the database
func migrationsTableExists(db models.DBService) bool {
	sql := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'migrations'
	);`

	// Use the Query method and manually handle the single row result
	rows, err := db.Query(sql)
	if err != nil {
		logger.Error("Error checking if migrations table exists", "error", err)
		return false
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Error("No result from migrations table existence check")
		return false
	}

	var exists bool
	err = rows.Scan(&exists)
	if err != nil {
		logger.Error("Error scanning migrations table existence result", "error", err)
		return false
	}

	return exists
}

func getMigrations(db models.DBService) []migration {
	sql := "select file_name, version, name, created_at from migrations order by version asc;"

	// Use the helper function that works with the DBService interface
	migrations, err := CollectRowsToStructFromService[migration](db, sql)
	if err != nil {
		logger.Error("Database Error while collecting migration versions", "error", err)
		return nil
	}

	return migrations
}

// updateMigrations writes a new row in the migrations table to record our action
func updateMigrations(db models.DBService, migration_files []string) {
	for _, f := range migration_files {
		migration := strings.Split(f, "_")
		sql :=
			`INSERT INTO migrations(file_name, version, name) 
			 VALUES($1, $2, $3);`
		result, err := db.ExecSQL(sql, f, migration[0], migration[1])
		if err != nil {
			logger.Error("Database Error while Inserting Migration Records", "error", err, "result", result)
		}
	}
}

func runCommand(command string, args ...string) ([]byte, error) {

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}
