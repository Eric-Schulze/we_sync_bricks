package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	DBConfig DBConfig `mapstructure:"db"`
}

func LoadConfig() (AppConfig, error) {
	// Set file name and type for environment configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()
	
	// Find and read the config file
	err := viper.ReadInConfig() 
	if err != nil { 
		// Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var dbConfig DBConfig
	err = viper.UnmarshalKey("db", &dbConfig)
	if err != nil {
		return AppConfig{}, err
	}

	return AppConfig{dbConfig}, nil
}


