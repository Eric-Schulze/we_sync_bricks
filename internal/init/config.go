package init

import (
	"github.com/spf13/viper"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

func LoadConfig() (models.AppConfig, error) {
	dbConfig, err := LoadDBConfig()
	if err != nil {
		return models.AppConfig{}, err
	}

	return models.AppConfig{DBConfig: dbConfig}, nil
}

// LoadDBConfig loads database configuration from config file
func LoadDBConfig() (models.DBConfig, error) {
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
